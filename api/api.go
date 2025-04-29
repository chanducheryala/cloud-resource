package api

import (
	"context"
	"encoding/json"
	"github.com/chanducheryala/cloud-resource/internal/analyzer"
	"github.com/chanducheryala/cloud-resource/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"net/http"
	"os"
	"strconv"
	"sync"
	"github.com/joho/godotenv"
)

var (
	resourceMutex sync.RWMutex
	resources     []models.CloudResource
	redisClient   *redis.Client
	logger        *zap.Logger
)

var (
	redisAddr string
	redisDB   int
)

func LoadAPIConfig() {
	_ = godotenv.Load()
	redisAddr = os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	redisDB = 0
	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		if v, err := strconv.Atoi(dbStr); err == nil {
			redisDB = v
		}
	}
}

func GetLogger() *zap.Logger {
	setupLogger()
	return logger
}

func setupLogger() {
	if logger == nil {
		l, err := zap.NewProduction()
		if err != nil {
			panic(err)
		}
		logger = l
	}
}

func setupRedis() {
	redisClient = redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   redisDB,
	})
	logger.Info("Redis client initialized")
}

func getAllResources(c *gin.Context) {
	resourceMutex.RLock()
	defer resourceMutex.RUnlock()
	logger.Info("GET /resources", zap.Int("count", len(resources)))
	c.JSON(http.StatusOK, resources)
}

func getResourceByID(c *gin.Context) {
	id := c.Param("id")
	resourceMutex.RLock()
	defer resourceMutex.RUnlock()
	for _, r := range resources {
		if r.GetId() == id {
			logger.Info("GET /resources/:id", zap.String("id", id))
			c.JSON(http.StatusOK, r)
			return
		}
	}
	logger.Warn("Resource not found", zap.String("id", id))
	c.JSON(http.StatusNotFound, gin.H{"error": "resource not found"})
}

func getResourceHistory(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	entries, err := redisClient.LRange(ctx, "resource:"+id+":history", 0, -1).Result()
	if err != nil {
		logger.Error("Redis LRange failed", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var history []map[string]interface{}
	for _, entry := range entries {
		var m map[string]interface{}
		json.Unmarshal([]byte(entry), &m)
		history = append(history, m)
	}
	logger.Info("GET /resources/:id/history", zap.String("id", id), zap.Int("entries", len(history)))
	c.JSON(http.StatusOK, history)
}

func saveResourceToRedis(res models.CloudResource) {
	ctx := context.Background()
	b, _ := json.Marshal(res)
	err := redisClient.RPush(ctx, "resource:"+res.GetId()+":history", b).Err()
	if err != nil {
		logger.Error("Redis RPush failed", zap.String("id", res.GetId()), zap.Error(err))
	}
}

func StartAPIServer(sharedResources *[]models.CloudResource, mutex *sync.RWMutex, suggestionSink analyzer.SuggestionSink, suggestionSinkType string) {
	LoadAPIConfig()
	resources = *sharedResources
	resourceMutex = *mutex
	setupLogger()
	logger.Info("Logger initialized")
	setupRedis()
	SetSuggestionSink(suggestionSink, suggestionSinkType)

	r := gin.Default()
	r.GET("/resources", getAllResources)
	r.GET("/resources/:id", getResourceByID)
	r.GET("/resources/:id/history", getResourceHistory)
	r.GET("/suggestions", getSuggestions)
	r.GET("/status", getStatus)
	go r.Run(":8080")
	logger.Info("API server started")
}
