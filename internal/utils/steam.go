package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"yt-api/internal/types"
)

var apiKey = os.Getenv("STEAM_API_KEY")

func GetProfileFromSteam(steamId string) (*types.ProfileResponse, error) {
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

	var profile types.ProfileResponse
	if err := json.Unmarshal(body, &profile); err != nil {
		log.Println("Error occurred while unmarshalling response body:", err)
		return nil, err
	}

	return &profile, nil
}
