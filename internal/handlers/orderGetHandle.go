package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func OrderHandler(c *gin.Context) {
	steamID, exist := c.Get("steamID")
	if exist {
		c.JSON(http.StatusOK, gin.H{
			"id": steamID,
		})
		return
	}
	c.AbortWithStatusJSON(401, gin.H{"error": "authentication required"})
}
