package logging

import (
	"log"
)

// Service for logger
type Service struct {
	enabled bool
}

// Enable logging
func (service *Service) Enable() {
	service.enabled = true
}

// Logf formats log message
func (service *Service) Logf(format string, params ...interface{}) {
	if service.enabled {
		log.Printf(format, params...)
	}
}

// Log message to log
func (service *Service) Log(msg string) {
	log.Print(msg)
}
