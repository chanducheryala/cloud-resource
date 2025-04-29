package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/chanducheryala/cloud-resource/api"
	"github.com/chanducheryala/cloud-resource/internal/analyzer"
	"github.com/chanducheryala/cloud-resource/internal/models"
	"github.com/chanducheryala/cloud-resource/utils"
)

func main() {
	resources := utils.GenerateMockResources()
	var resourceMutex sync.RWMutex
	out := make(chan models.CloudResource)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	suggestionSink := &analyzer.InMemorySuggestionSink{}
	suggestionSinkType := "memory"

	go api.StartAPIServer(&resources, &resourceMutex, suggestionSink, suggestionSinkType)

	logger := api.GetLogger()
	go utils.StartSimulation(ctx, resources, 1*time.Second, out, &resourceMutex, logger)

	go func() {
		<-ctx.Done()
		close(out)
	}()

	for res := range out {
		fmt.Println(res)
	}
}
