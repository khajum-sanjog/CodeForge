package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"codeforge/internal/kzm"
	"codeforge/internal/logger"

	"github.com/fsnotify/fsnotify"
)

// Watcher manages file system monitors (fsnotify) and remote repository polling.
type Watcher struct {
	daemon      *Daemon
	fsWatcher   *fsnotify.Watcher
	logger      *logger.Logger
	mu          sync.Mutex
	gitSHAs     map[string]string // repo_branch -> last commit SHA
	timers      map[string]*time.Timer // debounce timers for folder paths
	stopChan    chan struct{}
}

// NewWatcher instantiates a new Watcher service.
func NewWatcher(d *Daemon, l *logger.Logger) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &Watcher{
		daemon:    d,
		fsWatcher: fsw,
		logger:    l,
		gitSHAs:   make(map[string]string),
		timers:    make(map[string]*time.Timer),
		stopChan:  make(chan struct{}),
	}, nil
}

// Start initiates the watcher loop.
func (w *Watcher) Start(ctx context.Context) {
	// 1. Start FS Watcher loop
	go func() {
		for {
			select {
			case event, ok := <-w.fsWatcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
					// Check if this is a pipeline config update
					if strings.HasSuffix(event.Name, ".kzm") {
						w.logger.Log("daemon", "INFO", "Detected config change: %s. Reloading...", filepath.Base(event.Name))
						w.daemon.ReloadPipeline(event.Name)
						continue
					}

					// Otherwise, match against folder triggers
					w.handleFolderWrite(event.Name)
				}
			case err, ok := <-w.fsWatcher.Errors:
				if !ok {
					return
				}
				w.logger.Log("daemon", "ERROR", "FS Watcher error: %v", err)
			case <-ctx.Done():
				return
			case <-w.stopChan:
				return
			}
		}
	}()

	// 2. Start Repo Polling loop (every 30s)
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				w.pollGitRepos(ctx)
			case <-ctx.Done():
				return
			case <-w.stopChan:
				return
			}
		}
	}()
}

// Stop closes the file watcher.
func (w *Watcher) Stop() {
	close(w.stopChan)
	_ = w.fsWatcher.Close()
}

// WatchFolder adds a local directory path to the fsnotify watch list.
func (w *Watcher) WatchFolder(path string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.fsWatcher.Add(path)
}

// UnwatchFolder removes a local directory path from the fsnotify watch list.
func (w *Watcher) UnwatchFolder(path string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.fsWatcher.Remove(path)
}

func (w *Watcher) handleFolderWrite(path string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Find if any pipeline is watching this folder path
	// Debounce writes (500ms)
	w.daemon.mu.RLock()
	var projectToTrigger string
	var watchPath string
	for _, p := range w.daemon.pipelines {
		for _, trig := range p.Program.Triggers {
			if trig.Source == "folder" && strings.HasPrefix(path, trig.Path) {
				projectToTrigger = p.Program.Meta.Name
				watchPath = trig.Path
				break
			}
		}
		if projectToTrigger != "" {
			break
		}
	}
	w.daemon.mu.RUnlock()

	if projectToTrigger == "" {
		return
	}

	// Debounce logic
	if timer, ok := w.timers[watchPath]; ok {
		timer.Stop()
	}

	w.timers[watchPath] = time.AfterFunc(500*time.Millisecond, func() {
		w.logger.Log(projectToTrigger, "INFO", "Folder change detected. Firing trigger...")
		w.daemon.Trigger(projectToTrigger, "Folder change watcher")
	})
}

func (w *Watcher) pollGitRepos(ctx context.Context) {
	w.daemon.mu.RLock()
	pipelines := make([]*Pipeline, 0, len(w.daemon.pipelines))
	for _, p := range w.daemon.pipelines {
		pipelines = append(pipelines, p)
	}
	w.daemon.mu.RUnlock()

	for _, p := range pipelines {
		for _, trig := range p.Program.Triggers {
			if trig.Source == "github" || trig.Source == "gitlab" {
				go w.checkRepoUpdate(ctx, p.Program.Meta.Name, trig)
			}
		}
	}
}

func (w *Watcher) checkRepoUpdate(ctx context.Context, project string, trig *kzm.Trigger) {
	key := fmt.Sprintf("%s_%s_%s", trig.Source, trig.Repo, trig.Branch)
	
	var apiURL string
	if trig.Source == "github" {
		apiURL = fmt.Sprintf("https://api.github.com/repos/%s/commits/%s", trig.Repo, trig.Branch)
	} else { // gitlab
		// Escape repo path for gitlab
		escapedRepo := strings.ReplaceAll(trig.Repo, "/", "%2F")
		apiURL = fmt.Sprintf("https://gitlab.com/api/v4/projects/%s/repository/commits/%s", escapedRepo, trig.Branch)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return
	}

	// Try loading secret GITHUB_TOKEN or GITLAB_TOKEN if defined in process env
	if trig.Source == "github" {
		if token := os.Getenv("GITHUB_TOKEN"); token != "" {
			req.Header.Set("Authorization", "token "+token)
		}
	} else {
		if token := os.Getenv("GITLAB_TOKEN"); token != "" {
			req.Header.Set("PRIVATE-TOKEN", token)
		}
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var sha string
	if trig.Source == "github" {
		var ghCommit struct {
			SHA string `json:"sha"`
		}
		if err := json.Unmarshal(body, &ghCommit); err == nil {
			sha = ghCommit.SHA
		}
	} else { // gitlab
		var glCommit struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(body, &glCommit); err == nil {
			sha = glCommit.ID
		}
	}

	if sha == "" {
		return
	}

	w.mu.Lock()
	oldSHA, ok := w.gitSHAs[key]
	w.gitSHAs[key] = sha
	w.mu.Unlock()

	// If oldSHA exists and is different, trigger deploy
	if ok && oldSHA != sha {
		w.logger.Log(project, "INFO", "New commit detected on %s (%s): %s -> %s", trig.Source, trig.Branch, formatCommitSHA(oldSHA), formatCommitSHA(sha))
		w.daemon.TriggerWithSHA(project, fmt.Sprintf("%s push (%s)", trig.Source, trig.Branch), sha)
	}
}

func formatCommitSHA(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}
