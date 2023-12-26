package apksm

import (
	"sync"
	"time"
)

type ServiceStatusData struct {
	rwmu          sync.RWMutex
	ServiceStatus map[*Service][]*statusAtTime `json:"service-status"`
}

type statusAtTime struct {
	Time time.Time `json:"time"`
	// bool represent service online or offline
	Status bool `json:"running"`
}

func NewServiceStatusData(services Services) *ServiceStatusData {
	serviceStatusData := &ServiceStatusData{
		ServiceStatus: make(map[*Service][]*statusAtTime),
	}

	for _, service := range services {
		serviceStatusData.ServiceStatus[service] = make([]*statusAtTime, 0, 100)
	}

	return serviceStatusData
}

// SetStatusAtTimeForService updates map with new entry containing current time and service status at that time
func (s *ServiceStatusData) SetStatusAtTimeForService(service *Service, timeNow time.Time, status bool) {
	s.rwmu.Lock()
	defer s.rwmu.Unlock()
	s.ServiceStatus[service] = append(s.ServiceStatus[service], &statusAtTime{Time: timeNow, Status: status})
}

func (s *ServiceStatusData) GetServiceStatus() map[*Service][]*statusAtTime {
	s.rwmu.RLock()
	defer s.rwmu.RUnlock()
	return s.ServiceStatus
}
