package notification

import (
	"fmt"
	"os"
	"time"

	"github.com/windowmonitor/pkg/storage"
)

type Notifier struct {
	db     *storage.Storage
	fcmKey string
}

func NewNotifier(db *storage.Storage) *Notifier {
	return &Notifier{
		db:     db,
		fcmKey: os.Getenv("FCM_SERVER_KEY"),
	}
}

func (n *Notifier) CheckAndNotify() error {
	if n.fcmKey == "" {
		return fmt.Errorf("FCM server key not configured")
	}

	deviceToken := os.Getenv("FCM_DEVICE_TOKEN")
	if deviceToken == "" {
		return fmt.Errorf("FCM device token not configured")
	}

	currentHour := time.Now().Hour()

	if currentHour == 21 {
		return n.SendDailySummary(deviceToken)
	}

	return nil
}

func (n *Notifier) SendDailySummary(deviceToken string) error {
	stats, err := n.db.GetDailyStats()
	if err != nil {
		return fmt.Errorf("failed to get daily stats: %v", err)
	}

	fmt.Printf("Would send daily summary for %d window stats\n", len(stats))
	return nil
}
