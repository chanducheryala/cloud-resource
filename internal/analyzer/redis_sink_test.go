package analyzer

import (
	"context"
	"github.com/go-redis/redis/v8"
	"os"
	"testing"
	"time"
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
