package analyzer

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/chanducheryala/cloud-resource/internal/models"
)

func setupTestRedisSink(t *testing.T) *RedisSuggestionSink {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	client := redis.NewClient(&redis.Options{Addr: addr, DB: 1}) 
	if err := client.Ping(context.Background()).Err(); err != nil {
		t.Skip("Redis not available: " + err.Error())
	}
	sink := NewRedisSuggestionSink(client, "test-suggestions")
	_ = sink.ClearSuggestions()
	return sink
}

func TestRedisSuggestionSink_AddAndGet(t *testing.T) {
	t.Run("Lambda suggestion fields", func(t *testing.T) {
		sink := setupTestRedisSink(t)
		lambda := &models.Lambda{ID: "lambda-test", Invocations: 50, Errors: 4, CostPerMillion: 0.30, Owner: "Automation", LastModified: time.Now().Unix()}
		AnalyzeResource(lambda.GetUsage(), lambda, sink)
		suggestions := sink.GetSuggestions()
		for _, s := range suggestions {
			if s.ResourceID == "lambda-test" && s.ResourceType == "Lambda" {
				owner, ok := s.Details["owner"].(string)
				impact, ok2 := s.Details["business_impact"].(string)
				if !ok || owner == "" || !ok2 || impact == "" {
					t.Errorf("Lambda suggestion missing owner or business_impact: %+v", s.Details)
				}
			}
		}
	})
	t.Run("ELB suggestion fields", func(t *testing.T) {
		sink := setupTestRedisSink(t)
		elb := &models.ELB{ID: "elb-test", RequestCount: 500, HealthyHosts: 1, CostPerHour: 0.03, Owner: "WebOps", LastChecked: time.Now().Unix()}
		AnalyzeResource(elb.GetUsage(), elb, sink)
		suggestions := sink.GetSuggestions()
		for _, s := range suggestions {
			if s.ResourceID == "elb-test" && s.ResourceType == "ELB" {
				owner, ok := s.Details["owner"].(string)
				impact, ok2 := s.Details["business_impact"].(string)
				if !ok || owner == "" || !ok2 || impact == "" {
					t.Errorf("ELB suggestion missing owner or business_impact: %+v", s.Details)
				}
			}
		}
	})
	t.Run("S3 suggestion fields", func(t *testing.T) {
		sink := setupTestRedisSink(t)
		s3 := &models.S3{ID: "s3-test", UsedGB: 2000, ObjectCount: 2000000, CostPerGB: 0.04, Owner: "Backup", LastAccessed: time.Now().Unix()}
		AnalyzeResource(s3.GetUsage(), s3, sink)
		suggestions := sink.GetSuggestions()
		for _, s := range suggestions {
			if s.ResourceID == "s3-test" && s.ResourceType == "S3" {
				owner, ok := s.Details["owner"].(string)
				impact, ok2 := s.Details["business_impact"].(string)
				if !ok || owner == "" || !ok2 || impact == "" {
					t.Errorf("S3 suggestion missing owner or business_impact: %+v", s.Details)
				}
			}
		}
	})
	t.Run("DynamoDB suggestion fields", func(t *testing.T) {
		sink := setupTestRedisSink(t)
		db := &models.DynamoDB{ID: "ddb-test", ReadCapacity: 25, WriteCapacity: 25, ItemCount: 2000000, CostPerHr: 0.30, Owner: "Product", LastUpdated: time.Now().Unix()}
		AnalyzeResource(db.GetUsage(), db, sink)
		suggestions := sink.GetSuggestions()
		for _, s := range suggestions {
			if s.ResourceID == "ddb-test" && s.ResourceType == "DynamoDB" {
				owner, ok := s.Details["owner"].(string)
				impact, ok2 := s.Details["business_impact"].(string)
				if !ok || owner == "" || !ok2 || impact == "" {
					t.Errorf("DynamoDB suggestion missing owner or business_impact: %+v", s.Details)
				}
			}
		}
	})
	t.Run("Owner and Business Impact", func(t *testing.T) {
		sink := setupTestRedisSink(t)
		sug := Suggestion{
			ResourceID:   "vm-owner-test",
			ResourceType: "VM",
			Message:      "Test with owner and business impact",
			Timestamp:    time.Now(),
			Details: map[string]interface{}{
				"owner": "Finance Team",
				"business_impact": "Wasteful VM; immediate savings if terminated.",
			},
		}
		sink.AddSuggestion(sug)
		suggestions := sink.GetSuggestions()
		found := false
		for _, s := range suggestions {
			if s.ResourceID == "vm-owner-test" && s.Message == "Test with owner and business impact" {
				owner, ok := s.Details["owner"].(string)
				impact, ok2 := s.Details["business_impact"].(string)
				if ok && owner != "" && ok2 && impact != "" {
					found = true
				}
			}
		}
		if !found {
			t.Error("Owner or business_impact not set or empty in suggestion")
		}
	})

	sink := setupTestRedisSink(t)
	sug := Suggestion{
		ResourceID:   "vm-test",
		ResourceType: "VM",
		Message:      "Test suggestion",
		Timestamp:    time.Now(),
	}
	sink.AddSuggestion(sug)
	suggestions := sink.GetSuggestions()
	if len(suggestions) == 0 {
		t.Fatal("No suggestions returned from Redis")
	}
	found := false
	for _, s := range suggestions {
		if s.ResourceID == "vm-test" && s.Message == "Test suggestion" {
			found = true
		}
	}
	if !found {
		t.Error("Added suggestion not found in Redis")
	}
}

func TestRedisSuggestionSink_ClearSuggestions(t *testing.T) {
	sink := setupTestRedisSink(t)
	sug := Suggestion{
		ResourceID:   "vm-clear",
		ResourceType: "VM",
		Message:      "To be cleared",
		Timestamp:    time.Now(),
	}
	sink.AddSuggestion(sug)
	_ = sink.ClearSuggestions()
	suggestions := sink.GetSuggestions()
	if len(suggestions) != 0 {
		t.Error("Suggestions not cleared from Redis")
	}
}

func TestRedisSuggestionSink_ExpireSuggestions(t *testing.T) {
	sink := setupTestRedisSink(t)
	sug := Suggestion{
		ResourceID:   "vm-expire",
		ResourceType: "VM",
		Message:      "Expire me",
		Timestamp:    time.Now(),
	}
	sink.AddSuggestion(sug)
	err := sink.ExpireSuggestions(2 * time.Second)
	if err != nil {
		t.Fatalf("Failed to set expire: %v", err)
	}
	time.Sleep(3 * time.Second)
	suggestions := sink.GetSuggestions()
	if len(suggestions) != 0 {
		t.Error("Suggestions did not expire from Redis")
	}
}
