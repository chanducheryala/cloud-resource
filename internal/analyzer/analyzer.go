package analyzer

import (
	"github.com/chanducheryala/cloud-resource/internal/models"
	"sync"
	"time"
)

type Suggestion struct {
	ResourceID          string                 `json:"resource_id"`
	ResourceType        string                 `json:"resource_type"`
	Message             string                 `json:"message"`
	EstimatedSavingsUSD float64                `json:"estimated_savings_usd,omitempty"`
	Severity            string                 `json:"severity"`
	Timestamp           time.Time              `json:"timestamp"`
	Action              string                 `json:"action"`
	Details             map[string]interface{} `json:"details,omitempty"`
	DocsLink            string                 `json:"docs_link,omitempty"`
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
			if r.GetUsage() < 10.0 {
				sink.AddSuggestion(Suggestion{
					ResourceID:          r.GetId(),
					ResourceType:        "VM",
					Message:             "VM '" + r.GetId() + "' is underutilized (CPU < 10%) for 14 days. Consider resizing or terminating to eliminate waste.",
					EstimatedSavingsUSD: 45.00,
					Severity:            "Critical",
					Timestamp:           time.Now(),
					Action:              "Resize or terminate",
					Details: map[string]interface{}{
						"region":           "us-east-1",
						"owner":            "Team A",
						"current_type":     "t3.large",
						"recommended_type": "t3.small",
					},
					DocsLink: "https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-resize.html",
				})
			}
			if r.GetUsage() > 90.0 {
				sink.AddSuggestion(Suggestion{
					ResourceID:          r.GetId(),
					ResourceType:        "VM",
					Message:             "VM '" + r.GetId() + "' is over-provisioned. Consider rightsizing to reduce spend.",
					EstimatedSavingsUSD: 60.00,
					Severity:            "Info",
					Timestamp:           time.Now(),
					Action:              "Resize down",
					Details: map[string]interface{}{
						"current_type":     "t3.xlarge",
						"recommended_type": "t3.large",
					},
					DocsLink: "https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-resize.html",
				})
			}
			if r.CostPerHour > 0.5 {
				sink.AddSuggestion(Suggestion{
					ResourceID:          r.GetId(),
					ResourceType:        "VM",
					Message:             "VM '" + r.GetId() + "' has a high hourly cost. Consider moving to a reserved or spot instance.",
					EstimatedSavingsUSD: 100.00,
					Severity:            "Warning",
					Timestamp:           time.Now(),
					Action:              "Switch pricing model",
					Details: map[string]interface{}{
						"current_type": "expensive-type",
						"hourly_cost":  r.CostPerHour,
					},
					DocsLink: "https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-on-demand-reserved-instances.html",
				})
			}
		case *models.Storage:
			if r.UsedGB < 1.0 {
				sink.AddSuggestion(Suggestion{
					ResourceID:          r.ID,
					ResourceType:        "Storage",
					Message:             "Storage '" + r.ID + "' is idle and can be moved to a lower-cost storage class to eliminate waste.",
					EstimatedSavingsUSD: 10.00,
					Severity:            "Warning",
					Timestamp:           time.Now(),
					Action:              "Move to infrequent access tier",
					Details: map[string]interface{}{
						"region":        "us-east-1",
						"storage_class": "standard",
					},
					DocsLink: "https://docs.aws.amazon.com/AmazonS3/latest/userguide/storage-class-intro.html",
				})
			}
			if r.UsedGB > 900.0 {
				sink.AddSuggestion(Suggestion{
					ResourceID:          r.ID,
					ResourceType:        "Storage",
					Message:             "Storage '" + r.ID + "' is nearing capacity. Review and clean up unused data to avoid unnecessary expansion costs.",
					EstimatedSavingsUSD: 0.0,
					Severity:            "Critical",
					Timestamp:           time.Now(),
					Action:              "Cleanup or optimize",
					Details: map[string]interface{}{
						"used_gb": r.UsedGB,
					},
					DocsLink: "https://docs.aws.amazon.com/AmazonS3/latest/userguide/quotas.html",
				})
			}
			if time.Now().Unix() - r.LastAccessed > 90 * 24 * 3600 {
				sink.AddSuggestion(Suggestion{
					ResourceID:          r.ID,
					ResourceType:        "Storage",
					Message:             "Storage '" + r.ID + "' has not been accessed for 90+ days. Consider archiving or deleting.",
					EstimatedSavingsUSD: 20.00,
					Severity:            "Info",
					Timestamp:           time.Now(),
					Action:              "Archive or delete",
					Details: map[string]interface{}{
						"last_accessed": r.LastAccessed,
					},
					DocsLink: "https://docs.aws.amazon.com/AmazonS3/latest/userguide/object-lifecycle-mgmt.html",
				})
			}
			if r.CostPerGB > 0.10 {
				sink.AddSuggestion(Suggestion{
					ResourceID:          r.ID,
					ResourceType:        "Storage",
					Message:             "Storage '" + r.ID + "' has a high cost per GB. Consider moving to a lower-cost storage class.",
					EstimatedSavingsUSD: 15.00,
					Severity:            "Warning",
					Timestamp:           time.Now(),
					Action:              "Change storage class",
					Details: map[string]interface{}{
						"cost_per_gb": r.CostPerGB,
					},
					DocsLink: "https://docs.aws.amazon.com/AmazonS3/latest/userguide/storage-class-intro.html",
				})
			}
		case *models.Database:
			if r.Connections < 5 {
				sink.AddSuggestion(Suggestion{
					ResourceID:          r.ID,
					ResourceType:        "Database",
					Message:             "Database '" + r.ID + "' is over-provisioned. Consider downsizing to reduce waste.",
					EstimatedSavingsUSD: 25.00,
					Severity:            "Info",
					Timestamp:           time.Now(),
					Action:              "Downsize instance",
					Details: map[string]interface{}{
						"engine":           "Postgres",
						"current_size":     "db.m5.large",
						"recommended_size": "db.t3.medium",
					},
					DocsLink: "https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Overview.DBInstanceClass.html",
				})
			}
			if r.Connections > 150 {
				sink.AddSuggestion(Suggestion{
					ResourceID:          r.ID,
					ResourceType:        "Database",
					Message:             "Database '" + r.ID + "' has a high number of connections. Consider scaling up or load balancing.",
					EstimatedSavingsUSD: 0.0,
					Severity:            "Critical",
					Timestamp:           time.Now(),
					Action:              "Scale up or load balance",
					Details: map[string]interface{}{
						"connections": r.Connections,
					},
					DocsLink: "https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_WorkingWithConnections.html",
				})
			}
			if r.CPUUsage > 70.0 {
				sink.AddSuggestion(Suggestion{
					ResourceID:          r.ID,
					ResourceType:        "Database",
					Message:             "Database '" + r.ID + "' has high CPU usage. Consider query optimization or upgrading instance.",
					EstimatedSavingsUSD: 0.0,
					Severity:            "Warning",
					Timestamp:           time.Now(),
					Action:              "Optimize or upgrade",
					Details: map[string]interface{}{
						"cpu_usage": r.CPUUsage,
					},
					DocsLink: "https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/MonitoringOverview.html",
				})
			}
		}
	}()
}
