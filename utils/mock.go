package utils

import (
	"context"
	"github.com/chanducheryala/cloud-resource/internal/analyzer"
	"github.com/chanducheryala/cloud-resource/internal/models"
	"go.uber.org/zap"
	"time"
)

func StartSimulation(ctx context.Context, resources []models.CloudResource, interval time.Duration, out chan models.CloudResource, logger *zap.Logger, suggestionSink analyzer.SuggestionSink) {
	for _, resource := range resources {
		go func(res models.CloudResource) {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					res.UpdateUsage()
					analyzer.AnalyzeResource(res.GetUsage(), res, suggestionSink)
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
		&models.VM{ID: "vm-1", CostPerHour: 0.05, Owner: "Finance Team"},
		&models.VM{ID: "vm-2", CostPerHour: 0.10, Owner: "Engineering"},
		&models.Storage{ID: "s-1", CostPerGB: 0.02, LastAccessed: time.Now().Unix(), Owner: "Data Science"},
		&models.Database{ID: "db-1", CostPerHr: 0.20, Owner: "Analytics"},
		&models.S3{ID: "s3-1", UsedGB: 500, ObjectCount: 100000, CostPerGB: 0.023, Owner: "Backup", LastAccessed: time.Now().Unix()}, 
		&models.DynamoDB{ID: "ddb-1", ReadCapacity: 10, WriteCapacity: 5, ItemCount: 10000, CostPerHr: 0.10, Owner: "Product", LastUpdated: time.Now().Unix()}, 
		&models.Lambda{ID: "lambda-1", Invocations: 1000, Errors: 2, CostPerMillion: 0.20, Owner: "Automation", LastModified: time.Now().Unix()}, // Lambda (new struct)
		&models.ELB{ID: "elb-1", RequestCount: 50000, HealthyHosts: 3, CostPerHour: 0.025, Owner: "WebOps", LastChecked: time.Now().Unix()}, 
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
	case *models.Lambda:
		return r.ID + ", Lambda, Owner: " + r.Owner
	default:
		return "unknown resource"
	}
}
