package gui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

// setupSystemTray initializes the system tray icon and menu controls.
func (a *CodeForgeApp) setupSystemTray() {
	if deskApp, ok := a.FyneApp.(desktop.App); ok {
		logo := LoadLogo()
		deskApp.SetSystemTrayIcon(logo)

		// Set Double-Click handler (if supported by OS, otherwise menu items work)
		// Fyne DesktopApp doesn't always expose double-click callback directly on all platforms,
		// but standard menus are fully cross-platform.

		a.refreshSystemTrayMenu(deskApp)
	}
}

func (a *CodeForgeApp) refreshSystemTrayMenu(deskApp desktop.App) {
	_, running := a.getDaemonPid()

	// 1. Daemon Status label
	statusLabel := "○ Daemon Stopped"
	if running {
		statusLabel = "● Daemon Running"
	}

	// 2. Show Dashboard action
	showItem := fyne.NewMenuItem("Show Dashboard", func() {
		a.MainWindow.Show()
		a.MainWindow.RequestFocus()
	})

	// 3. Dynamic pipeline triggers
	pipelinesDir := filepath.Join(secretsResolvePath("~/.codeforge"), "pipelines")
	files, _ := filepath.Glob(filepath.Join(pipelinesDir, "*.kzm"))
	
	pipelineItems := []*fyne.MenuItem{}
	for _, file := range files {
		project := strings.TrimSuffix(filepath.Base(file), ".kzm")
		pipelineItems = append(pipelineItems, fyne.NewMenuItem(fmt.Sprintf("%s → Run Now", project), func() {
			_ = a.Daemon.Trigger(project, "system tray click")
		}))
	}

	// 4. Daemon controls
	var stopStartItem *fyne.MenuItem
	if running {
		stopStartItem = fyne.NewMenuItem("Stop Daemon", func() {
			a.Daemon.Stop()
			fyne.Do(func() {
				a.refreshSystemTrayMenu(deskApp)
				a.refreshSidebar()
			})
		})
	} else {
		stopStartItem = fyne.NewMenuItem("Start Daemon", func() {
			_ = a.Daemon.Start()
			fyne.Do(func() {
				a.refreshSystemTrayMenu(deskApp)
				a.refreshSidebar()
			})
		})
	}

	restartItem := fyne.NewMenuItem("Restart Daemon", func() {
		a.Daemon.Stop()
		_ = a.Daemon.Start()
		fyne.Do(func() {
			a.refreshSystemTrayMenu(deskApp)
			a.refreshSidebar()
		})
	})

	quitItem := fyne.NewMenuItem("Quit CodeForge", func() {
		a.Daemon.Stop()
		a.FyneApp.Quit()
	})

	// Compile menu
	menuItems := []*fyne.MenuItem{
		fyne.NewMenuItem("CodeForge", nil),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem(statusLabel, nil),
		fyne.NewMenuItemSeparator(),
		showItem,
		fyne.NewMenuItemSeparator(),
	}

	// Append pipelines
	if len(pipelineItems) > 0 {
		for _, pi := range pipelineItems {
			menuItems = append(menuItems, pi)
		}
		menuItems = append(menuItems, fyne.NewMenuItemSeparator())
	}

	menuItems = append(menuItems, stopStartItem, restartItem, fyne.NewMenuItemSeparator(), quitItem)

	menu := fyne.NewMenu("CodeForge Tray", menuItems...)
	deskApp.SetSystemTrayMenu(menu)
}

func secretsResolvePath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}
