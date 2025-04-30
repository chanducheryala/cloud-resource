package analyzer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisSuggestionSink struct {
	Client *redis.Client
	Key    string
}

func NewRedisSuggestionSink(client *redis.Client, key string) *RedisSuggestionSink {
	return &RedisSuggestionSink{Client: client, Key: key}
}

func (r *RedisSuggestionSink) AddSuggestion(sug Suggestion) {
	ctx := context.Background()
	b, _ := json.Marshal(sug)

	exists, _ := r.Client.Exists(ctx, r.Key).Result()
	if exists == 0 {
		_, _ = r.Client.Do(ctx, "JSON.SET", r.Key, ".", "[]").Result()
	}

	fmt.Println("Adding suggestion:", string(b))
	jsonStr, err := r.Client.Do(ctx, "JSON.GET", r.Key, ".").Result()
	var suggestions []Suggestion
	if err == nil && jsonStr != nil {
		bs, ok := jsonStr.(string)
		if ok {
			_ = json.Unmarshal([]byte(bs), &suggestions)
		}
	}
	for _, existing := range suggestions {
		if existing.ResourceID == sug.ResourceID &&
			existing.Action == sug.Action &&
			existing.ResourceType == sug.ResourceType &&
			existing.Message == sug.Message {
			return
		}
	}
	_, _ = r.Client.Do(ctx, "JSON.ARRAPPEND", r.Key, ".", string(b)).Result()
}

func (r *RedisSuggestionSink) GetSuggestions() []Suggestion {
	ctx := context.Background()
	var result []Suggestion
	jsonStr, err := r.Client.Do(ctx, "JSON.GET", r.Key, ".").Result()
	if err != nil || jsonStr == nil {
		return nil
	}
	bs, ok := jsonStr.(string)
	if !ok {
		return nil
	}
	_ = json.Unmarshal([]byte(bs), &result)
	return result
}

func (r *RedisSuggestionSink) ClearSuggestions() error {
	ctx := context.Background()
	_, err := r.Client.Do(ctx, "JSON.SET", r.Key, ".", "[]").Result()
	return err
}

func (r *RedisSuggestionSink) ExpireSuggestions(ttl time.Duration) error {
	ctx := context.Background()
	return r.Client.Expire(ctx, r.Key, ttl).Err()
}
