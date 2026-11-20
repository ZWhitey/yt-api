package handlers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// HealthHandler returns the service liveness status.
func HealthHandler(c *gin.Context) {
	version := os.Getenv("VERSION")
	if version == "" {
		version = "unknown"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"version": version,
	})
}
