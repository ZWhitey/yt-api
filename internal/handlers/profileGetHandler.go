package handlers

import (
	"net/http"
	"os"
	"yt-api/internal/utils"

	"github.com/gin-gonic/gin"
)

var apiKey = os.Getenv("STEAM_API_KEY")

func GetProfileHandler(c *gin.Context) {
	steamID, exist := c.Get("steamID")
	if !exist {
		c.AbortWithStatusJSON(401, gin.H{"error": "authentication required"})
		return
	}

	profile, err := utils.GetProfileFromSteam(steamID.(string))
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"name":    profile.Response.Players[0].PersonaName,
		"steamId": steamID,
		"avatar":  profile.Response.Players[0].Avatar,
	})
}
