package apksm

import (
	"apksm/notify"
)

type Settings struct {
	Monitor       *MonitorSettings
	Notifications *NotificationSettings
}

type MonitorSettings struct {
	CheckInterval             int  `json:"checkInterval"`
	MonitorSystemUsage        bool `json:"monitorSystemUsage"`
	CPUMax                    int  `json:"CPUMax"`
	MemoryMax                 int  `json:"memoryMax"`
	ExponentialBackoffSeconds int  `json:"exponentialBackoffSeconds"`
}

type NotificationSettings struct {
	Email    []*notify.EmailSettings    `json:"email"`
	Telegram []*notify.TelegramSettings `json:"telegram"`
	Webhook  []*notify.WebhookSettings  `json:"webhook"`
}

func (n *NotificationSettings) GetNotifiers() (notifiers notify.Notifiers) {
	for _, email := range n.Email {
		emailNotifier := &notify.EmailNotifier{Settings: email}
		notifiers = append(notifiers, emailNotifier)
	}
	for _, telegram := range n.Telegram {
		telegramNotifier := &notify.TelegramNotifier{Settings: telegram}
		notifiers = append(notifiers, telegramNotifier)
	}
	for _, webhook := range n.Webhook {
		webhookNotifier := &notify.WebhookNotifier{Settings: webhook}
		notifiers = append(notifiers, webhookNotifier)
	}
	return
}
