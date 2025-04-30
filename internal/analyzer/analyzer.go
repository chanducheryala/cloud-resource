package analyzer

import (
	"sync"
	"time"
	"github.com/chanducheryala/cloud-resource/internal/models"
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
	ClearSuggestions() error
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

func (s *InMemorySuggestionSink) ClearSuggestions() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.suggestions = []Suggestion{}
	return nil
}

func AnalyzeResource(usage float64, resource models.CloudResource, sink SuggestionSink) {
	go func() {
		switch r := resource.(type) {	
		case *models.VM:
			if r.GetUsage() < 100.0 {
				sink.AddSuggestion(Suggestion{
					ResourceID:   r.GetId(),
					ResourceType: "VM",
					Message:      "Underutilized VM (CPU < 100%)",
					Timestamp:    time.Now(),
				})
			}
		case *models.Storage:
			if true {
				sink.AddSuggestion(Suggestion{
					ResourceID:   r.ID,
					ResourceType: "Storage",
					Message:      "Idle storage volume (LastAccessed > 0d)",
					Timestamp:    time.Now(),
				})
			}
		case *models.Database:
			if true {
				sink.AddSuggestion(Suggestion{
					ResourceID:   r.ID,
					ResourceType: "Database",
					Message:      "Over-provisioned database (Connections > 0)",
					Timestamp:    time.Now(),
				})
			}
		}
	}()
}
