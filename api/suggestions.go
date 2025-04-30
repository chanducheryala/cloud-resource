package api

import (
	"github.com/chanducheryala/cloud-resource/internal/analyzer"
	"github.com/gin-gonic/gin"
	"net/http"
)

var suggestionSink analyzer.SuggestionSink
var suggestionSinkType string

func SetSuggestionSink(sink analyzer.SuggestionSink, sinkType string) {
	suggestionSink = sink
	suggestionSinkType = sinkType
}

func getSuggestions(c *gin.Context) {
	if suggestionSink == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "suggestion sink not configured"})
		return
	}
	suggestions := suggestionSink.GetSuggestions()
	c.JSON(http.StatusOK, suggestions)
}

func clearSuggestions(c *gin.Context) {
	if suggestionSink == nil {
		c.JSON(500, gin.H{"error": "suggestion sink not configured"})
		return
	}
	err := suggestionSink.ClearSuggestions()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "Suggestions cleared"})
}

func getStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"sink": suggestionSinkType})
}
