package apksm

import (
	"apksm/logger"
	"apksm/notify"
	"apksm/track"
	"fmt"
	"os"
	"regexp"
	"time"
)

type Status struct {
	Ok bool
}

type Monitor struct {
	// Holds settings and services
	config *Config

	// Channel used to schedule checks for services
	checkerCh chan *Service

	// Notification methods used to send messages when service can't be reached
	notifiers notify.Notifiers

	// Channel used for receive services that couldn't be reached
	notifierChannel chan *Service

	// To reduce notification spam, tracker is used to delay notifications
	notificationTracker map[*Service]*track.TimeTracker

	// Sending to stop channel makes program exit
	stop chan struct{}

	// TODO: For each service, keep map with time and up/down status
	serviceStatusData *ServiceStatusData
}

func NewMonitor(c *Config) *Monitor {
	m := &Monitor{
		config:              c,
		checkerCh:           make(chan *Service),
		notifiers:           c.Settings.Notifications.GetNotifiers(),
		notifierChannel:     make(chan *Service),
		notificationTracker: make(map[*Service]*track.TimeTracker),
		stop:                make(chan struct{}),
		serviceStatusData:   NewServiceStatusData(c.Services),
	}
	m.initialize()
	return m
}

func (m *Monitor) initialize() {
	// Initialize notification methods to reduce overhead
	for _, notifier := range m.notifiers {
		if initializer, ok := notifier.(notify.Initializer); ok {
			logger.Logln("Initializing", initializer)
			initializer.Initialize()
		}
	}

	for _, service := range m.config.Services {
		// Initialize notificationTracker
		m.notificationTracker[service] = NewTrackerWithExpBackoff(m.config.Settings.Monitor.ExponentialBackoffSeconds)

		// Set default CheckInterval and Timeout for services who miss them
		switch {
		case service.CheckInterval <= 0:
			service.CheckInterval = m.config.Settings.Monitor.CheckInterval
		}
	}
}

// NewTrackerWithExpBackoff creates TimeTracker with ExpBackoff as Delayer
func NewTrackerWithExpBackoff(expBackoffSeconds int) *track.TimeTracker {
	return track.NewTracker(track.NewExpBackoff(expBackoffSeconds))
}

// Run runs monitor infinitely
func (m *Monitor) Run() {
	m.RunForSeconds(0)
}

// RunForSeconds runs monitor for runningSeconds seconds or infinitely if 0 is passed as an argument
func (m *Monitor) RunForSeconds(runningSeconds int) {
	if runningSeconds != 0 {
		go func() {
			runningSecondsTime := time.Duration(runningSeconds) * time.Second
			<-time.After(runningSecondsTime)
			m.stop <- struct{}{}
		}()
	}

	for _, service := range m.config.Services {
		// Schedules each service to run every specific seconds
		go m.scheduleService(service)
	}

	logger.Logln("Starting APK Service Monitor ...")
	m.monitor()
}

func (m *Monitor) scheduleService(s *Service) {
	// Initial
	m.checkerCh <- s

	// Periodic
	tickerSeconds := time.NewTicker(time.Duration(s.CheckInterval) * time.Second)
	for range tickerSeconds.C {
		m.checkerCh <- s
	}
}

func (m *Monitor) monitor() {
	go m.listenForChecks()
	go m.checkSystemUsage()
	go m.listenForNotifications()

	// Wait for termination signal then exit monitor
	<-m.stop
	logger.Logln("Terminating.")
	os.Exit(0)
}

func (m *Monitor) listenForChecks() {
	for service := range m.checkerCh {
		m.checkServiceStatus(service)
	}
}

func (m *Monitor) listenForNotifications() {
	for service := range m.notifierChannel {
		timeTracker := m.notificationTracker[service]
		if timeTracker.IsReady() {
			nextDelay, nextTime := timeTracker.SetNext()
			logger.Logln("Sending notifications for", service)
			go m.notifiers.NotifyAll(fmt.Sprintf("%s (%s)", service.Name, service))
			logger.Logln("Next available notification for", service.String(), "in", nextDelay, "at", nextTime)
		}
	}
}

func (m *Monitor) checkServiceStatus(service *Service) {
	go func() {
		logger.Logln("Checking", service)
		serviceStatus := Status{}
		if service.IsRunning() {
			serviceStatus.Ok = true
		} else {
			serviceStatus.Ok = false
		}
		// Check user given regex with service logs
		regex, _ := regexp.MatchString(service.SpecificPattern, service.Logs())
		if regex {
			st := fmt.Sprintf("Found `%s` pattern in %s service logs", service.SpecificPattern, service)
			logger.Logln(st)
			go m.notifiers.NotifyAll(st)
		}
		// Register the current status to calculate the uptime
		m.serviceStatusData.SetStatusAtTimeForService(service, time.Now(), serviceStatus.Ok)

		// Handle error
		if !serviceStatus.Ok {
			logger.Logln("ERROR", service)
			if service.SaveLogsOnStop {
				logger.Logln("Saving", service, "logs")
				service.SaveLogs()
			}
			if service.RestartIfDown {
				// If user specified that this service must restart if it is down

				logger.Logln("Restart if down is ON")
				logger.Logln("RESTARTING", service)

				if service.Restart() {
					// Try to restart
					logger.Logln("RESTARTED", service)
				} else {
					logger.Logln("ERROR RESTARTING", service)
				}
			} else {
				logger.Logln("Restart if down is OFF")
			}
			go func() {
				m.notifierChannel <- service
			}()
			return
		}

		// Handle success
		logger.Logln("OK", service)
		// Reset time tracker for service
		if m.notificationTracker[service].HasBeenRan() {
			m.notificationTracker[service] = NewTrackerWithExpBackoff(m.config.Settings.Monitor.ExponentialBackoffSeconds)
		}
	}()
}

func (m *Monitor) checkSystemUsage() {
	if m.config.Settings.Monitor.MonitorSystemUsage {
		// System usage monitoring is enabled
		ticker := time.NewTicker(time.Duration(m.config.Settings.Monitor.CheckInterval) * time.Second)
		quit := make(chan struct{})
		go func() {
			for {
				select {
				case <-ticker.C:
					idle0, total0 := getCPUStatus()
					memory := getMemoryStats()
					time.Sleep(3 * time.Second)
					idle1, total1 := getCPUStatus()

					idleTicks := float64(idle1 - idle0)
					totalTicks := float64(total1 - total0)
					cpuUsage := 100 * (totalTicks - idleTicks) / totalTicks

					if (memory.MemTotal - memory.MemAvailable) > m.config.Settings.Monitor.MemoryMax {
						st := fmt.Sprintf("Memory usage gone upper than %dkB!", m.config.Settings.Monitor.MemoryMax)
						logger.Logln(st)
						go m.notifiers.NotifyAll(st)
					}
					if cpuUsage > float64(m.config.Settings.Monitor.CPUMax) {
						st := fmt.Sprintf("CPU usage gone upper than %d percent!", m.config.Settings.Monitor.CPUMax)
						logger.Logln(st)
						go m.notifiers.NotifyAll(st)
					}
				case <-quit:
					ticker.Stop()
					return
				}
			}
		}()
	}
}
