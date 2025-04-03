package notification

import (
	"fmt"
	"time"
	"unsafe"

	"github.com/windowmonitor/pkg/storage"
	"golang.org/x/sys/windows"
)

// Windows notification constants
const (
	NIM_ADD     = 0x00000000
	NIM_MODIFY  = 0x00000001
	NIM_DELETE  = 0x00000002
	NIF_MESSAGE = 0x00000001
	NIF_ICON    = 0x00000002
	NIF_TIP     = 0x00000004
	NIF_INFO    = 0x00000010
	NIIF_INFO   = 0x00000001
	WM_APP      = 0x8000
	WM_TRAYICON = WM_APP + 1
)

// NOTIFYICONDATA structure for Windows API
type NOTIFYICONDATA struct {
	CbSize           uint32
	HWnd             windows.Handle
	UID              uint32
	UFlags           uint32
	UCallbackMessage uint32
	HIcon            windows.Handle
	SzTip            [128]uint16
	DwState          uint32
	DwStateMask      uint32
	SzInfo           [256]uint16
	UVersion         uint32
	SzInfoTitle      [64]uint16
	DwInfoFlags      uint32
	GuidItem         windows.GUID
}

// WindowsNotifier handles Windows native notifications
type WindowsNotifier struct {
	db              *storage.Storage
	shell32         *windows.LazyDLL
	shellNotifyIcon *windows.LazyProc
	lastWindow      string
	lastDuration    time.Duration
}

// NewWindowsNotifier creates a new Windows notifier
func NewWindowsNotifier(db *storage.Storage) *WindowsNotifier {
	shell32 := windows.NewLazyDLL("shell32.dll")
	shellNotifyIcon := shell32.NewProc("Shell_NotifyIconW")

	return &WindowsNotifier{
		db:              db,
		shell32:         shell32,
		shellNotifyIcon: shellNotifyIcon,
	}
}

// ShowWindowSwitchNotification shows a notification when switching between windows
func (wn *WindowsNotifier) ShowWindowSwitchNotification(windowTitle string, duration time.Duration) error {
	// Only show notification if the duration is significant (more than 5 seconds)
	if duration.Seconds() < 5 {
		return nil
	}

	// Format the notification message
	message := fmt.Sprintf("You spent %s on %s", formatDuration(duration), windowTitle)

	// Show the notification using Windows API
	return wn.showNotification("Window Monitor", message)
}

// ShowSummaryNotification shows a summary notification when the application stops
func (wn *WindowsNotifier) ShowSummaryNotification() error {
	stats, err := wn.db.GetDailyStats()
	if err != nil {
		return fmt.Errorf("failed to get daily stats: %v", err)
	}

	if len(stats) == 0 {
		return nil
	}

	// Find the most used window
	var mostUsedWindow string
	var longestDuration time.Duration

	for _, stat := range stats {
		if stat.Duration > longestDuration {
			longestDuration = stat.Duration
			mostUsedWindow = stat.Title
		}
	}

	// Format the notification message
	message := fmt.Sprintf("Today's summary: Most used window was %s (%s)",
		mostUsedWindow, formatDuration(longestDuration))

	// Show the notification using Windows API
	return wn.showNotification("Window Monitor Summary", message)
}

// showNotification displays a Windows notification
func (wn *WindowsNotifier) showNotification(title, message string) error {
	// Print to console for logging purposes
	fmt.Printf("[NOTIFICATION] %s: %s\n", title, message)

	// Get a handle to the foreground window to make notification more visible
	user32 := windows.NewLazyDLL("user32.dll")
	getForegroundWindow := user32.NewProc("GetForegroundWindow")
	hwnd, _, _ := getForegroundWindow.Call()

	// Convert strings to UTF16
	titleUTF16, _ := windows.UTF16FromString(title)
	messageUTF16, _ := windows.UTF16FromString(message)

	// Create notification data structure with proper flags for Windows notification panel
	nid := NOTIFYICONDATA{
		CbSize:           uint32(unsafe.Sizeof(NOTIFYICONDATA{})),
		HWnd:             windows.Handle(hwnd), // Use foreground window handle
		UID:              1,                    // Use a consistent ID for this application
		UFlags:           NIF_INFO | NIF_TIP | NIF_MESSAGE,
		UCallbackMessage: WM_TRAYICON,
		DwInfoFlags:      NIIF_INFO,
		UVersion:         4, // NOTIFYICON_VERSION_4 for Windows 7 and later
	}

	// Copy the title and message to the notification structure
	copy(nid.SzInfoTitle[:], titleUTF16)
	copy(nid.SzInfo[:], messageUTF16)

	// Call the Windows API
	ret, _, _ := wn.shellNotifyIcon.Call(
		uintptr(NIM_ADD),
		uintptr(unsafe.Pointer(&nid)),
	)

	// Check if the call was successful
	if ret == 0 {
		return fmt.Errorf("failed to show notification")
	}

	// Remove the notification after a short delay
	go func() {
		time.Sleep(5 * time.Second)
		wn.shellNotifyIcon.Call(
			uintptr(NIM_DELETE),
			uintptr(unsafe.Pointer(&nid)),
		)
	}()

	return nil
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d.Hours() >= 1 {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		return fmt.Sprintf("%dh %dm", hours, minutes)
	} else if d.Minutes() >= 1 {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	} else {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
}
