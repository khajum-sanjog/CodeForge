package executor

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Snapshot represents a packaged deployment state.
type Snapshot struct {
	Project   string            `json:"project"`
	Timestamp string            `json:"timestamp"`
	Files     map[string]string `json:"files"` // relPath -> base64(content)
}

// SaveSnapshot walks the target sourceDir, encodes files into base64, and writes a JSON snapshot.
func SaveSnapshot(project, sourceDir, snapshotDir string) (string, error) {
	resolvedDir := resolvePath(snapshotDir)
	projectDir := filepath.Join(resolvedDir, sanitizeFilename(project))
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return "", err
	}

	snap := Snapshot{
		Project:   project,
		Timestamp: time.Now().Format(time.RFC3339),
		Files:     make(map[string]string),
	}

	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		// Skip skipped paths
		if shouldSkip(rel) {
			return nil
		}

		// Read and encode file
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		data, err := io.ReadAll(f)
		if err != nil {
			return err
		}

		snap.Files[rel] = base64.StdEncoding.EncodeToString(data)
		return nil
	})
	if err != nil {
		return "", err
	}

	// Marshall to JSON
	snapData, err := json.Marshal(snap)
	if err != nil {
		return "", err
	}

	// Write JSON file
	filename := fmt.Sprintf("%d.json", time.Now().UnixNano())
	snapPath := filepath.Join(projectDir, filename)
	if err := os.WriteFile(snapPath, snapData, 0644); err != nil {
		return "", err
	}

	return snapPath, nil
}

// RestoreSnapshot reads a JSON snapshot and writes its files into the target outputDir.
func RestoreSnapshot(snapshotPath, outputDir string) error {
	data, err := os.ReadFile(snapshotPath)
	if err != nil {
		return err
	}

	var snap Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return err
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	for rel, b64 := range snap.Files {
		fileData, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			return err
		}

		filePath := filepath.Join(outputDir, rel)
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return err
		}

		if err := os.WriteFile(filePath, fileData, 0644); err != nil {
			return err
		}
	}

	return nil
}

// GetLatestSnapshot finds and returns the path to the newest snapshot for the project.
func GetLatestSnapshot(project, snapshotDir string) (string, error) {
	resolvedDir := filepath.Join(resolvePath(snapshotDir), sanitizeFilename(project))
	matches, err := filepath.Glob(filepath.Join(resolvedDir, "*.json"))
	if err != nil || len(matches) == 0 {
		return "", fmt.Errorf("no snapshots found for project %q", project)
	}

	// Sort matches
	sort.Strings(matches)
	return matches[len(matches)-1], nil
}

// PruneSnapshots keeps only the last N snapshot files, deleting any older versions.
func PruneSnapshots(project, snapshotDir string, keepLast int) {
	if keepLast <= 0 {
		return
	}
	resolvedDir := filepath.Join(resolvePath(snapshotDir), sanitizeFilename(project))
	matches, err := filepath.Glob(filepath.Join(resolvedDir, "*.json"))
	if err != nil || len(matches) <= keepLast {
		return
	}

	sort.Strings(matches)
	toDelete := len(matches) - keepLast
	for i := 0; i < toDelete; i++ {
		_ = os.Remove(matches[i])
	}
}

func resolvePath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

func sanitizeFilename(name string) string {
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, name)
}

func shouldSkip(relPath string) bool {
	parts := strings.Split(relPath, string(filepath.Separator))
	for _, p := range parts {
		if p == ".git" || p == "node_modules" || p == "vendor" || p == ".kzm" || p == ".codeforge" {
			return true
		}
	}
	return false
}
