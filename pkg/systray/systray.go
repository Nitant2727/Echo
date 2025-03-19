package systray

import (
	"fmt"
	"os"

	"github.com/getlantern/systray"
	"github.com/windowmonitor/pkg/analytics"
	"github.com/windowmonitor/pkg/storage"
)

type TrayManager struct {
	storage    *storage.Storage
	visualizer *analytics.Visualizer
}

func NewTrayManager(storage *storage.Storage, visualizer *analytics.Visualizer) *TrayManager {
	return &TrayManager{
		storage:    storage,
		visualizer: visualizer,
	}
}

func (tm *TrayManager) Start() {
	systray.Run(tm.onReady, tm.onExit)
}

func (tm *TrayManager) onReady() {
	systray.SetIcon(getIcon())
	systray.SetTitle("Window Monitor")
	systray.SetTooltip("Window Monitor - Track your window usage")

	mOpenStats := systray.AddMenuItem("Open Statistics", "View your window usage statistics")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Exit Window Monitor")

	go func() {
		for {
			select {
			case <-mOpenStats.ClickedCh:
				// Open browser to statistics page
				fmt.Println("Opening statistics dashboard...")
				// The visualizer is already running on :8080

			case <-mQuit.ClickedCh:
				fmt.Println("Exiting...")
				systray.Quit()
				return
			}
		}
	}()
}

func (tm *TrayManager) onExit() {
	// Cleanup
	os.Exit(0)
}

// getIcon returns a simple icon for the system tray
func getIcon() []byte {
	// This is a simple 16x16 icon
	return []byte{
		0x00, 0x00, 0x00, 0x00,
		0x00, 0xFF, 0xFF, 0x00,
		0x00, 0xFF, 0xFF, 0x00,
		0x00, 0xFF, 0xFF, 0x00,
		0x00, 0xFF, 0xFF, 0x00,
		0x00, 0xFF, 0xFF, 0x00,
		0x00, 0x00, 0x00, 0x00,
	}
}