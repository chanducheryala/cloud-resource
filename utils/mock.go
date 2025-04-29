package utils

import (
	"context"
	"github.com/chanducheryala/cloud-resource/internal/models"
	"go.uber.org/zap"
	"sync"
	"time"
)

func StartSimulation(ctx context.Context, resources []models.CloudResource, interval time.Duration, out chan models.CloudResource, mutex *sync.RWMutex, logger *zap.Logger) {
	for _, resource := range resources {
		go func(res models.CloudResource) {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					res.UpdateUsage()
					mutex.Lock()
					mutex.Unlock()
					logger.Info("Resource state", zap.String("resource", resourceToString(res)))
					out <- res
					time.Sleep(interval)
				}
			}
		}(resource)
	}
}

func GenerateMockResources() []models.CloudResource {
	return []models.CloudResource{
		&models.VM{ID: "vm-1", CostPerHour: 0.05},
		&models.VM{ID: "vm-2", CostPerHour: 0.10},
		&models.Storage{ID: "s-1", CostPerGB: 0.02, LastAccessed: time.Now().Unix()},
		&models.Database{ID: "db-1", CostPerHr: 0.20},
	}
}

func resourceToString(res models.CloudResource) string {
	switch r := res.(type) {
	case *models.VM:
		return r.String()
	case *models.Storage:
		return r.String()
	case *models.Database:
		return r.String()
	default:
		return "unknown resource"
	}
}
