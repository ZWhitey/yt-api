package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	. "yt-api/internal/types"

	"github.com/gin-gonic/gin"
)

var apiKey = os.Getenv("STEAM_API_KEY")

func GetProfileHandler(c *gin.Context) {
	steamID, exist := c.Get("steamID")
	if !exist {
		c.AbortWithStatusJSON(401, gin.H{"error": "authentication required"})
		return
	}

	profile, err := getProfileFromSteam(steamID.(string))
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

func getProfileFromSteam(steamId string) (*ProfileResponse, error) {
	url := fmt.Sprintf("https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002/?key=%s&steamids=%s", apiKey, steamId)
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Error occurred while getting profile from steam:", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error occurred while reading response body:", err)
		return nil, err
	}

	var profile ProfileResponse
	if err := json.Unmarshal(body, &profile); err != nil {
		log.Println("Error occurred while unmarshalling response body:", err)
		return nil, err
	}

	return &profile, nil
}
