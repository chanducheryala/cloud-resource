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
	Priority            int                    `json:"priority"` // 1=Critical, 2=Warning, 3=Info (lower is higher priority)
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
		case *models.Lambda:
			totalInvocations := r.Invocations
			errors := r.Errors
			errorRate := 0.0
			if totalInvocations > 0 {
				errorRate = float64(errors) / float64(totalInvocations)
			}
			if errorRate > 0.05 {
				details := map[string]interface{}{
					"owner": r.Owner,
					"error_rate": errorRate,
					"invocations": totalInvocations,
					"errors": errors,
				}
				details["business_impact"] = "High error rates may indicate wasted compute and lost business logic."
				sink.AddSuggestion(Suggestion{
					ResourceID:   r.GetId(),
					ResourceType: "Lambda",
					Message:      "Lambda '" + r.GetId() + "' has a high error rate (>5%). Investigate and fix failing invocations.",
					EstimatedSavingsUSD: 0.0,
					Severity:     "Warning",
					Priority:     2,
					Timestamp:    time.Now(),
					Action:       "Debug and fix errors",
					Details:      details,
					DocsLink:     "https://docs.aws.amazon.com/lambda/latest/dg/invocation-retries.html",
				})
			}
			if totalInvocations < 100 {
				details := map[string]interface{}{
					"owner": r.Owner,
					"invocations": totalInvocations,
				}
				details["business_impact"] = "Idle Lambda functions can be removed to reduce clutter and potential attack surface."
				sink.AddSuggestion(Suggestion{
					ResourceID:   r.GetId(),
					ResourceType: "Lambda",
					Message:      "Lambda '" + r.GetId() + "' has low invocation rates (<100/month). Consider removing or consolidating idle functions.",
					EstimatedSavingsUSD: 0.0,
					Severity:     "Info",
					Priority:     3,
					Timestamp:    time.Now(),
					Action:       "Review for removal",
					Details:      details,
					DocsLink:     "https://docs.aws.amazon.com/lambda/latest/dg/best-practices.html",
				})
			}

			if r.CostPerMillion > 0.25 {
				details := map[string]interface{}{
					"owner": r.Owner,
					"cost_per_million": r.CostPerMillion,
				}
				details["business_impact"] = "High Lambda costs may indicate inefficient code or configuration."
				sink.AddSuggestion(Suggestion{
					ResourceID:   r.GetId(),
					ResourceType: "Lambda",
					Message:      "Lambda '" + r.GetId() + "' has a high cost per million invocations (>$0.25). Review function configuration and usage.",
					EstimatedSavingsUSD: 5.0,
					Severity:     "Warning",
					Priority:     2,
					Timestamp:    time.Now(),
					Action:       "Optimize configuration",
					Details:      details,
					DocsLink:     "https://docs.aws.amazon.com/lambda/latest/dg/configuration-memory.html",
				})
			}


		case *models.DynamoDB:
			if r.ReadCapacity > 20 || r.WriteCapacity > 20 {
				details := map[string]interface{}{
					"owner": r.Owner,
					"read_capacity": r.ReadCapacity,
					"write_capacity": r.WriteCapacity,
				}
				details["business_impact"] = "Overprovisioned tables waste money on unused throughput."
				sink.AddSuggestion(Suggestion{
					ResourceID:   r.GetId(),
					ResourceType: "DynamoDB",
					Message:      "DynamoDB table '" + r.GetId() + "' is overprovisioned (Read/Write Capacity > 20). Consider scaling down provisioned throughput.",
					EstimatedSavingsUSD: r.CostPerHr * 24 * 30 * 0.5, 
					Severity:     "Warning",
					Priority:     2,
					Timestamp:    time.Now(),
					Action:       "Scale down provisioned throughput",
					Details:      details,
					DocsLink:     "https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/ProvisionedThroughput.html",
				})
			}
			if r.ItemCount > 1000000 {
				details := map[string]interface{}{
					"owner": r.Owner,
					"item_count": r.ItemCount,
				}
				details["business_impact"] = "Large tables may contain stale or unnecessary data, increasing costs."
				sink.AddSuggestion(Suggestion{
					ResourceID:   r.GetId(),
					ResourceType: "DynamoDB",
					Message:      "DynamoDB table '" + r.GetId() + "' is large (>1 million items). Review for archiving or partitioning.",
					EstimatedSavingsUSD: r.CostPerHr * 24 * 30 * 0.2, // assume 20% savings possible
					Severity:     "Info",
					Priority:     3,
					Timestamp:    time.Now(),
					Action:       "Review for archiving/partitioning",
					Details:      details,
					DocsLink:     "https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/best-practices.html",
				})
			}

			if r.CostPerHr > 0.25 {
				details := map[string]interface{}{
					"owner": r.Owner,
					"cost_per_hr": r.CostPerHr,
				}
				details["business_impact"] = "High DynamoDB costs may indicate overprovisioning or inefficient access patterns."
				sink.AddSuggestion(Suggestion{
					ResourceID:   r.GetId(),
					ResourceType: "DynamoDB",
					Message:      "DynamoDB table '" + r.GetId() + "' has a high cost per hour (>$0.25). Review usage and optimize table settings.",
					EstimatedSavingsUSD: (r.CostPerHr - 0.10) * 24 * 30,
					Severity:     "Warning",
					Priority:     2,
					Timestamp:    time.Now(),
					Action:       "Optimize table settings",
					Details:      details,
					DocsLink:     "https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/HowItWorks.ReadWriteCapacityMode.html",
				})
			}


		case *models.S3:
			if r.UsedGB > 1000 {
				details := map[string]interface{}{
					"owner": r.Owner,
					"used_gb": r.UsedGB,
				}
				details["business_impact"] = "Large S3 buckets may contain stale or unnecessary data, increasing costs."
				sink.AddSuggestion(Suggestion{
					ResourceID:   r.GetId(),
					ResourceType: "S3",
					Message:      "S3 Bucket '" + r.GetId() + "' is large (>1000 GB). Review for data lifecycle and retention policies.",
					EstimatedSavingsUSD: r.UsedGB * r.CostPerGB * 0.2, // assume 20% savings possible
					Severity:     "Warning",
					Priority:     2,
					Timestamp:    time.Now(),
					Action:       "Review and clean up old data",
					Details:      details,
					DocsLink:     "https://docs.aws.amazon.com/AmazonS3/latest/userguide/object-lifecycle-mgmt.html",
				})
			}
			if r.ObjectCount > 1000000 {
				details := map[string]interface{}{
					"owner": r.Owner,
					"object_count": r.ObjectCount,
				}
				details["business_impact"] = "Buckets with too many objects can increase management overhead and costs."
				sink.AddSuggestion(Suggestion{
					ResourceID:   r.GetId(),
					ResourceType: "S3",
					Message:      "S3 Bucket '" + r.GetId() + "' has more than 1 million objects. Consider consolidation or archiving.",
					EstimatedSavingsUSD: r.UsedGB * r.CostPerGB * 0.1, // assume 10% savings possible
					Severity:     "Info",
					Priority:     3,
					Timestamp:    time.Now(),
					Action:       "Consolidate or archive objects",
					Details:      details,
					DocsLink:     "https://docs.aws.amazon.com/AmazonS3/latest/userguide/optimizing-performance.html",
				})
			}

			if r.CostPerGB > 0.03 {
				details := map[string]interface{}{
					"owner": r.Owner,
					"cost_per_gb": r.CostPerGB,
				}
				details["business_impact"] = "High S3 cost per GB may indicate inefficient storage class selection."
				sink.AddSuggestion(Suggestion{
					ResourceID:   r.GetId(),
					ResourceType: "S3",
					Message:      "S3 Bucket '" + r.GetId() + "' has a high cost per GB (>$0.03). Review storage class and region.",
					EstimatedSavingsUSD: r.UsedGB * (r.CostPerGB - 0.023),
					Severity:     "Warning",
					Priority:     2,
					Timestamp:    time.Now(),
					Action:       "Review storage class",
					Details:      details,
					DocsLink:     "https://docs.aws.amazon.com/AmazonS3/latest/userguide/storage-class-intro.html",
				})
			}


		case *models.ELB:
			if r.RequestCount < 1000 {
				details := map[string]interface{}{
					"owner": r.Owner,
					"request_count": r.RequestCount,
				}
				details["business_impact"] = "Underutilized ELBs incur ongoing costs with minimal value."
				sink.AddSuggestion(Suggestion{
					ResourceID:   r.GetId(),
					ResourceType: "ELB",
					Message:      "ELB '" + r.GetId() + "' is underutilized (<1000 requests). Consider downsizing or removal.",
					EstimatedSavingsUSD: r.CostPerHour * 24 * 30,
					Severity:     "Info",
					Priority:     3,
					Timestamp:    time.Now(),
					Action:       "Review for downsizing/removal",
					Details:      details,
					DocsLink:     "https://docs.aws.amazon.com/elasticloadbalancing/latest/userguide/load-balancer-troubleshooting.html",
				})
			}

			if r.HealthyHosts < 2 {
				details := map[string]interface{}{
					"owner": r.Owner,
					"healthy_hosts": r.HealthyHosts,
				}
				details["business_impact"] = "Unhealthy ELBs may cause downtime or lost revenue."
				sink.AddSuggestion(Suggestion{
					ResourceID:   r.GetId(),
					ResourceType: "ELB",
					Message:      "ELB '" + r.GetId() + "' has fewer than 2 healthy hosts. Investigate target group health.",
					EstimatedSavingsUSD: 0.0,
					Severity:     "Warning",
					Priority:     2,
					Timestamp:    time.Now(),
					Action:       "Investigate health",
					Details:      details,
					DocsLink:     "https://docs.aws.amazon.com/elasticloadbalancing/latest/userguide/target-group-health-checks.html",
				})
			}
			costPerRequest := 0.0
			if r.RequestCount > 0 {
				costPerRequest = r.CostPerHour / float64(r.RequestCount)
			}
			if costPerRequest > 0.00005 {
				details := map[string]interface{}{
					"owner": r.Owner,
					"cost_per_request": costPerRequest,
				}
				details["business_impact"] = "High ELB cost per request may indicate over-provisioning or low traffic."
				sink.AddSuggestion(Suggestion{
					ResourceID:   r.GetId(),
					ResourceType: "ELB",
					Message:      "ELB '" + r.GetId() + "' has a high cost per request (>$0.00005). Review configuration and traffic patterns.",
					EstimatedSavingsUSD: r.CostPerHour * 24 * 30,
					Severity:     "Warning",
					Priority:     2,
					Timestamp:    time.Now(),
					Action:       "Optimize configuration",
					Details:      details,
					DocsLink:     "https://docs.aws.amazon.com/elasticloadbalancing/latest/userguide/load-balancer-cost-optimization.html",
				})
			}


		case *models.VM:
			if r.PreviousCostPerHour > 0 && r.CostPerHour > r.PreviousCostPerHour*1.5 {
				sink.AddSuggestion(Suggestion{
					ResourceID:          r.GetId(),
					ResourceType:        "VM",
					Message:             "Cost spike detected for VM '" + r.GetId() + "'. Hourly cost increased by more than 50%. Investigate recent changes or usage.",
					EstimatedSavingsUSD: (r.CostPerHour - r.PreviousCostPerHour) * 24 * 30,
					Severity:            "Critical",
					Priority:            1,
					Timestamp:           time.Now(),
					Action:              "Investigate cost anomaly",
					Details: map[string]interface{}{
						"previous_cost_per_hour": r.PreviousCostPerHour,
						"current_cost_per_hour": r.CostPerHour,
						"owner": r.Owner,
						"business_impact": "Sudden cost increase; investigate to prevent unexpected spend.",
					},
					DocsLink: "https://docs.aws.amazon.com/cost-management/latest/userguide/cost-anomaly-detection.html",
				})
			}
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
						"owner":            r.Owner,
						"current_type":     "t3.large",
						"recommended_type": "t3.small",
						"business_impact":  "No recent activity; freeing this VM will save significant costs.",
					},
					DocsLink: "https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-resize.html",
				})
			}
			if r.LastActive > 0 && time.Now().Unix()-r.LastActive > 30*24*3600 {
				sink.AddSuggestion(Suggestion{
					ResourceID:          r.GetId(),
					ResourceType:        "VM",
					Message:             "VM '" + r.GetId() + "' has not been active for 30+ days. Consider terminating to eliminate waste.",
					EstimatedSavingsUSD: r.CostPerHour * 24 * 30, // monthly
					Severity:            "Critical",
					Timestamp:           time.Now(),
					Action:              "Terminate",
					Details: map[string]interface{}{
						"owner":           r.Owner,
						"last_active":     r.LastActive,
						"business_impact": "Resource idle for over a month; terminating will save $/month.",
					},
					DocsLink: "https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/stop-start-instance.html",
				})
			}
			if r.GetUsage() > 90.0 {
				sink.AddSuggestion(Suggestion{
					ResourceID:          r.GetId(),
					ResourceType:        "VM",
					Message:             "VM '" + r.GetId() + "' is over-provisioned. Consider rightsizing to reduce spend.",
					EstimatedSavingsUSD: 60.00,
					Severity:            "Info",
					Priority:            3,
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
					Priority:            2,
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
			if r.PreviousCostPerGB > 0 && r.CostPerGB > r.PreviousCostPerGB*1.5 {
				sink.AddSuggestion(Suggestion{
					ResourceID:          r.ID,
					ResourceType:        "Storage",
					Message:             "Cost spike detected for Storage '" + r.ID + "'. Cost per GB increased by more than 50%. Investigate recent changes or storage class.",
					EstimatedSavingsUSD: (r.CostPerGB - r.PreviousCostPerGB) * r.UsedGB,
					Severity:            "Critical",
					Priority:            1,
					Timestamp:           time.Now(),
					Action:              "Investigate storage cost anomaly",
					Details: map[string]interface{}{
						"previous_cost_per_gb": r.PreviousCostPerGB,
						"current_cost_per_gb": r.CostPerGB,
						"owner": r.Owner,
						"business_impact": "Sudden storage cost increase; investigate to prevent unexpected spend.",
					},
					DocsLink: "https://docs.aws.amazon.com/cost-management/latest/userguide/cost-anomaly-detection.html",
				})
			}
			if r.UsedGB < 1.0 {
				sink.AddSuggestion(Suggestion{
					ResourceID:          r.ID,
					ResourceType:        "Storage",
					Message:             "Storage '" + r.ID + "' is idle and can be moved to a lower-cost storage class to eliminate waste.",
					EstimatedSavingsUSD: 10.00,
					Severity:            "Warning",
					Priority:            2,
					Timestamp:           time.Now(),
					Action:              "Move to infrequent access tier",
					Details: map[string]interface{}{
						"region":        "us-east-1",
						"storage_class": "standard",
						"owner":         r.Owner,
						"business_impact": "Idle storage can be archived or deleted to save costs.",
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
						"owner":   r.Owner,
						"business_impact": "Storage nearing capacity; cleaning up prevents additional spend.",
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
					Priority:            3,
					Timestamp:           time.Now(),
					Action:              "Archive or delete",
					Details: map[string]interface{}{
						"last_accessed": r.LastAccessed,
						"owner":         r.Owner,
						"business_impact": "No access in 90+ days; archiving can save costs.",
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
					Priority:            2,
					Timestamp:           time.Now(),
					Action:              "Change storage class",
					Details: map[string]interface{}{
						"cost_per_gb": r.CostPerGB,
						"owner":       r.Owner,
						"business_impact": "High storage cost; move to lower-cost class for efficiency.",
					},
					DocsLink: "https://docs.aws.amazon.com/AmazonS3/latest/userguide/storage-class-intro.html",
				})
			}
		case *models.Database:
			if r.PreviousCostPerHr > 0 && r.CostPerHr > r.PreviousCostPerHr*1.5 {
				sink.AddSuggestion(Suggestion{
					ResourceID:          r.ID,
					ResourceType:        "Database",
					Message:             "Cost spike detected for Database '" + r.ID + "'. Hourly cost increased by more than 50%. Investigate recent changes or usage.",
					EstimatedSavingsUSD: (r.CostPerHr - r.PreviousCostPerHr) * 24 * 30,
					Severity:            "Critical",
					Priority:            1,
					Timestamp:           time.Now(),
					Action:              "Investigate database cost anomaly",
					Details: map[string]interface{}{
						"previous_cost_per_hr": r.PreviousCostPerHr,
						"current_cost_per_hr": r.CostPerHr,
						"owner": r.Owner,
						"business_impact": "Sudden database cost increase; investigate to prevent unexpected spend.",
					},
					DocsLink: "https://docs.aws.amazon.com/cost-management/latest/userguide/cost-anomaly-detection.html",
				})
			}
			if r.Connections < 5 {
				sink.AddSuggestion(Suggestion{
					ResourceID:          r.ID,
					ResourceType:        "Database",
					Message:             "Database '" + r.ID + "' is over-provisioned. Consider downsizing to reduce waste.",
					EstimatedSavingsUSD: 25.00,
					Severity:            "Info",
					Priority:            3,
					Timestamp:           time.Now(),
					Action:              "Downsize instance",
					Details: map[string]interface{}{
						"engine":           "Postgres",
						"current_size":     "db.m5.large",
						"recommended_size": "db.t3.medium",
						"owner":            r.Owner,
						"business_impact":  "Over-provisioned DB; downsizing will reduce waste and save costs.",
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
						"owner":       r.Owner,
						"business_impact": "High connection count; scaling or balancing can prevent outages and improve business continuity.",
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
					Priority:            2,
					Timestamp:           time.Now(),
					Action:              "Optimize or upgrade",
					Details: map[string]interface{}{
						"cpu_usage": r.CPUUsage,
						"owner":     r.Owner,
						"business_impact": "High DB CPU usage; optimizing or upgrading can improve performance and user experience.",
					},
					DocsLink: "https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/MonitoringOverview.html",
				})
			}
		}
	}()
}
