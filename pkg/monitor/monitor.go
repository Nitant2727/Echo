package monitor

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"

	"github.com/windowmonitor/pkg/notification"
	"github.com/windowmonitor/pkg/storage"
	"golang.org/x/sys/windows"
)

type WindowMonitor struct {
	db *storage.Storage
	notifier *notification.WindowsNotifier
	lastWindow string
	lastTime time.Time
}

var (
	user32                = windows.NewLazyDLL("user32.dll")
	getForegroundWindow   = user32.NewProc("GetForegroundWindow")
	getWindowTextW        = user32.NewProc("GetWindowTextW")
)

func NewWindowMonitor(db *storage.Storage) *WindowMonitor {
	notifier := notification.NewWindowsNotifier(db)
	return &WindowMonitor{
		db: db,
		notifier: notifier,
		lastTime: time.Now(),
	}
}

func (w *WindowMonitor) getActiveWindowTitle() (string, error) {
	hwnd, _, _ := getForegroundWindow.Call()
	if hwnd == 0 {
		return "", fmt.Errorf("no active window")
	}

	buf := make([]uint16, 256)
	_, _, _ = getWindowTextW.Call(
		hwnd,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
	)

	return syscall.UTF16ToString(buf), nil
}

func (w *WindowMonitor) Start() {
	for {
		title, err := w.getActiveWindowTitle()
		if err == nil && title != "" {
			if title != w.lastWindow && w.lastWindow != "" {
				duration := time.Since(w.lastTime)
				if err := w.db.SaveWindowStats(w.lastWindow, duration); err != nil {
					fmt.Printf("Error saving window stats: %v\n", err)
				}
				
				// Show notification about the time spent on the previous window
				if err := w.notifier.ShowWindowSwitchNotification(w.lastWindow, duration); err != nil {
					fmt.Printf("Error showing notification: %v\n", err)
				}
				
				w.lastTime = time.Now()
			} else if w.lastWindow == "" {
				w.lastTime = time.Now()
			}
			w.lastWindow = title
		}
		time.Sleep(100 * time.Millisecond)
	}
}