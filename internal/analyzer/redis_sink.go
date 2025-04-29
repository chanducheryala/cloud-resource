package analyzer

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"time"
)

type RedisSuggestionSink struct {
	Client *redis.Client
	Key    string // e.g. "suggestions"
}

func NewRedisSuggestionSink(client *redis.Client, key string) *RedisSuggestionSink {
	return &RedisSuggestionSink{Client: client, Key: key}
}

func (r *RedisSuggestionSink) AddSuggestion(sug Suggestion) {
	ctx := context.Background()
	b, _ := json.Marshal(sug)
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
