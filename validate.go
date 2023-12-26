package apksm

import (
	"fmt"
)

func (c *Config) Validate() error {
	if err := c.Settings.Validate(); err != nil {
		return fmt.Errorf("invalid settings: %v", err)
	}
	if err := c.Services.Validate(); err != nil {
		return fmt.Errorf("invalid services: %v", err)
	}
	return nil
}

func (s *Settings) Validate() error {
	if err := s.Monitor.Validate(); err != nil {
		return fmt.Errorf("invalid monitor settings: %v", err)
	}
	if err := s.Notifications.Validate(); err != nil {
		return fmt.Errorf("invalid notification settings: %v", err)
	}
	return nil
}

func (ms *MonitorSettings) Validate() error {
	// ExponentialBackoffSeconds can be 0, which means when calculated, delay for notifications will always be 1 second

	if ms.CheckInterval <= 0 || ms.ExponentialBackoffSeconds < 0 {
		return fmt.Errorf("monitor settings missing")
	}
	return nil
}

func (ns *NotificationSettings) Validate() error {
	for _, email := range ns.Email {
		if err := email.Validate(); err != nil {
			return fmt.Errorf("invalid email settings: %v", err)
		}
	}
	for _, telegram := range ns.Telegram {
		if err := telegram.Validate(); err != nil {
			return fmt.Errorf("invalid telegram settings: %v", err)
		}
	}
	for _, webhook := range ns.Webhook {
		if err := webhook.Validate(); err != nil {
			return fmt.Errorf("invalid webhook settings: %v", err)
		}
	}
	return nil
}

func (services Services) Validate() error {
	if len(services) == 0 {
		return fmt.Errorf("no services found in config")
	}

	for _, service := range services {
		if err := service.Validate(); err != nil {
			return fmt.Errorf("invalid service settings: %s", err)
		}

	}
	return nil
}

func (s *Service) Validate() error {
	errServiceProperty := func(property string) error {
		return fmt.Errorf("missing service property %s", property)
	}
	switch {
	case s.Name == "":
		return errServiceProperty("name")
	}
	return nil
}
