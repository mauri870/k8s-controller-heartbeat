package k8sheartbeat

import (
	"context"
)

// HealthChecker defines a health check interface
type HealthChecker interface {
	// HealthCheck checks if a given component is healthy
	HealthCheck(context.Context, string) error
	// Ping checks if the health checker is running properly
	Ping() error
}
