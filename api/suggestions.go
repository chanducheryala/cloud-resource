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

func getStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"sink": suggestionSinkType})
}
