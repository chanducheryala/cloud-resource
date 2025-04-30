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
	"time"
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
	
	err := godotenv.Load()
	if err != nil {
		logger.Error("Failed to load environment variables", zap.Error(err))
	}

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
	l, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	logger = l
}

func setupRedis() {
	redisClient = redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   redisDB,
	})

	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		logger.Error("Failed to connect to Redis", zap.Error(err))
		return
	}
	logger.Info("Redis client initialized")
}

func getAllResources(c *gin.Context) {
	resourceMutex.RLock()
	defer resourceMutex.RUnlock()
	logger.Info("all resources", zap.Int("count", len(resources)))
	c.JSON(http.StatusOK, resources)
}

func getResourceByID(c *gin.Context) {
	id := c.Param("id")
	resourceMutex.RLock()
	defer resourceMutex.RUnlock()
	for _, r := range resources {
		if r.GetId() == id {
			logger.Info("resource by id", zap.String("id", id))
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
	entries, err := redisClient.LRange(ctx, "resource:"+ id +":history", 0, -1).Result()
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
	logger.Info("resources history for id", zap.String("id", id), zap.Int("entries", len(history)))
	c.JSON(http.StatusOK, history)
}

func GetRedisClient() *redis.Client {
	if redisClient == nil {
		setupRedis()
	}
	return redisClient
}

func StartAPIServer(ctx context.Context, sharedResources *[]models.CloudResource, suggestionSink analyzer.SuggestionSink, suggestionSinkType string) *http.Server {
	LoadAPIConfig()
	resources = *sharedResources	
	setupLogger()
	logger.Info("Logger initialized")
	
	setupRedis()
	
	SetSuggestionSink(suggestionSink, suggestionSinkType)

	r := gin.Default()
	
	r.GET("/resources", getAllResources)
	r.GET("/resources/:id", getResourceByID)
	r.GET("/resources/:id/history", getResourceHistory)
	r.GET("/suggestions", getSuggestions)
	r.POST("/suggestions/clear", clearSuggestions)
	r.GET("/status", getStatus)

	httpServer := &http.Server{
        Addr:    ":8080",
        Handler: r,
    }

    go func() {
        if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.Error("API server error", zap.Error(err))
        }
    }()
    logger.Info("API server started")

    go func() {
        <-ctx.Done()
        logger.Info("Shutting down API server...")
        shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        if err := httpServer.Shutdown(shutdownCtx); err != nil {
            logger.Error("API server forced to shutdown", zap.Error(err))
        }
    }()

    return httpServer
}
