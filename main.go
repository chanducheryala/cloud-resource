package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	"github.com/chanducheryala/cloud-resource/api"
	"github.com/chanducheryala/cloud-resource/internal/analyzer"
	"github.com/chanducheryala/cloud-resource/internal/models"
	"github.com/chanducheryala/cloud-resource/utils"
	"go.uber.org/zap"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	resources := utils.GenerateMockResources()

	out := make(chan models.CloudResource)

	logger := api.GetLogger()

	redisClient := api.GetRedisClient() 
	redisSink := analyzer.NewRedisSuggestionSink(redisClient, "suggestions")
	

	suggestionSinkType := "redis"

	server := api.StartAPIServer(ctx, &resources, redisSink, suggestionSinkType)

	go utils.StartSimulation(ctx, resources, 1 * time.Second, out, logger, redisSink)

	go func() {
		for res := range out {
			fmt.Println(res)
		}
	}()

	<-ctx.Done()
	
	logger.Info("Shutdown signal received, shutting down gracefully...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}
	
	logger.Info("Server exited")
}
