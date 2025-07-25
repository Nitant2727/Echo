package systray

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/getlantern/systray"
	"github.com/windowmonitor/pkg/analytics"
	"github.com/windowmonitor/pkg/notification"
	"github.com/windowmonitor/pkg/storage"
)

type TrayManager struct {
	storage    *storage.Storage
	visualizer *analytics.Visualizer
	notifier   *notification.WindowsNotifier
}

func NewTrayManager(storage *storage.Storage, visualizer *analytics.Visualizer) *TrayManager {
	// Create a Windows notifier
	winNotifier := notification.NewWindowsNotifier(storage)

	return &TrayManager{
		storage:    storage,
		visualizer: visualizer,
		notifier:   winNotifier,
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
				fmt.Println("Opening statistics dashboard...")
				url := "http://localhost:8080"
				err := exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
				if err != nil {
					fmt.Printf("Failed to open browser: %v\n", err)
				}

			case <-mQuit.ClickedCh:
				fmt.Println("Exiting...")
				systray.Quit()
				return
			}
		}
	}()
}

func (tm *TrayManager) onExit() {
	if err := tm.notifier.ShowSummaryNotification(); err != nil {
		fmt.Printf("Failed to show summary notification: %v\n", err)
	}

	// Cleanup and exit
	os.Exit(0)
}

// getIcon returns a simple icon for the system tray
func getIcon() []byte {
	// This is a properly formatted 16x16 icon in RGBA format
	icon := make([]byte, 16*16*4)
	
	// Fill with a simple blue square pattern
	for i := 0; i < 16; i++ {
		for j := 0; j < 16; j++ {
			pos := (i*16 + j) * 4
			// Create a blue color with full opacity
			icon[pos] = 0      // R
			icon[pos+1] = 0    // G
			icon[pos+2] = 255  // B
			icon[pos+3] = 255  // A (opacity)
		}
	}
	
	return icon
}
