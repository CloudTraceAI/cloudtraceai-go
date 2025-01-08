package cloudtrace

import "log"

type ServiceRegistry struct {
	Services []MonitoringService
}

type MonitoringService interface {
	Register() error
	Track(event string, data map[string]interface{}) error
	Monitor() error
}

func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{Services: []MonitoringService{}}
}

func (sr *ServiceRegistry) AddService(service MonitoringService) {
	sr.Services = append(sr.Services, service)
}

func (sr *ServiceRegistry) Initialize() error {
	for _, service := range sr.Services {
		if err := service.Register(); err != nil {
			return err
		}
	}
	log.Println("All services registered successfully.")
	return nil
}

func (sr *ServiceRegistry) TrackEvent(event string, data map[string]interface{}) error {
	for _, service := range sr.Services {
		if err := service.Track(event, data); err != nil {
			return err
		}
	}
	return nil
}

func (sr *ServiceRegistry) StartMonitoring() error {
	for _, service := range sr.Services {
		if err := service.Monitor(); err != nil {
			return err
		}
	}
	return nil
}
