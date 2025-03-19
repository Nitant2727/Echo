package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/windowmonitor/pkg/analytics"
	"github.com/windowmonitor/pkg/monitor"
	"github.com/windowmonitor/pkg/notification"
	"github.com/windowmonitor/pkg/storage"
	"github.com/windowmonitor/pkg/systray"
)

func main() {
	// Setup data directory in user's home
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get home directory: %v", err)
	}
	dataDir := filepath.Join(homeDir, ".windowmonitor")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	// Initialize the storage with SQLite database
	db, err := storage.NewStorage(filepath.Join(dataDir, "window_stats.db"))
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer db.Close()

	// Initialize components
	monitor := monitor.NewWindowMonitor(db)
	visualizer := analytics.NewVisualizer(db)
	notifier := notification.NewNotifier(db)
	trayManager := systray.NewTrayManager(db, visualizer)

	// Start the visualization server
	go func() {
		if err := visualizer.StartServer(":8080"); err != nil {
			log.Printf("Failed to start visualization server: %v", err)
		}
	}()

	// Start the monitoring process
	go monitor.Start()

	// Start the notification checker
	go func() {
		for {
			if err := notifier.CheckAndNotify(); err != nil {
				log.Printf("Notification error: %v", err)
			}
			time.Sleep(5 * time.Minute)
		}
	}()

	// Start system tray
	fmt.Println("Starting Window Monitor...")
	fmt.Println("View analytics dashboard at http://localhost:8080")
	fmt.Println("The application will run in the system tray")

	// Run the system tray (this blocks)
	trayManager.Start()
}