package analyzer

import (
	"context"
	"github.com/chanducheryala/cloud-resource/internal/models"
	"sync"
	"time"
)

type Suggestion struct {
	ResourceID   string    `json:"resource_id"`
	ResourceType string    `json:"resource_type"`
	Message      string    `json:"message"`
	Timestamp    time.Time `json:"timestamp"`
}

type SuggestionSink interface {
	AddSuggestion(s Suggestion)
	GetSuggestions() []Suggestion
}

type InMemorySuggestionSink struct {
	mu          sync.RWMutex
	suggestions []Suggestion
}

func (s *InMemorySuggestionSink) AddSuggestion(sug Suggestion) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.suggestions = append(s.suggestions, sug)
}

func (s *InMemorySuggestionSink) GetSuggestions() []Suggestion {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]Suggestion(nil), s.suggestions...)
}

func StartAnalyzer(ctx context.Context, resources *[]models.CloudResource, mutex *sync.RWMutex, sink SuggestionSink, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				mutex.RLock()
				for _, res := range *resources {
					switch r := res.(type) {
					case *models.VM:
						if r.CPUUsage < 10.0 {
							sink.AddSuggestion(Suggestion{
								ResourceID:   r.ID,
								ResourceType: "VM",
								Message:      "Underutilized VM (CPU < 10%)",
								Timestamp:    time.Now(),
							})
						}
					case *models.Storage:
						if time.Since(time.Unix(r.LastAccessed, 0)) > 30*24*time.Hour {
							sink.AddSuggestion(Suggestion{
								ResourceID:   r.ID,
								ResourceType: "Storage",
								Message:      "Idle storage volume (LastAccessed > 30d)",
								Timestamp:    time.Now(),
							})
						}
					case *models.Database:
						if r.Connections > 100 {
							sink.AddSuggestion(Suggestion{
								ResourceID:   r.ID,
								ResourceType: "Database",
								Message:      "Over-provisioned database (Connections > 100)",
								Timestamp:    time.Now(),
							})
						}
					}
				}
				mutex.RUnlock()
			}
		}
	}()
}
