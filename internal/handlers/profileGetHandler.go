package handlers

import (
	"context"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
	"yt-api/internal/model"
	"yt-api/internal/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var apiKey = os.Getenv("STEAM_API_KEY")
var emailPattern = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
var taiwanMobilePattern = regexp.MustCompile(`^09\d{8}$`)

type userProfile struct {
	RealName string `bson:"realName"`
	Email    string `bson:"email"`
	Phone    string `bson:"phone"`
}

type updateUserProfileRequest struct {
	RealName string `json:"realName" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Phone    string `json:"phone" binding:"required"`
}

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

	profileData := userProfile{}
	collection := model.Db.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = collection.FindOne(ctx, bson.M{"SteamID": steamID.(string)}).Decode(&profileData)
	if err != nil && err != mongo.ErrNoDocuments {
		log.Println("Error occurred while finding user profile:", err)
		c.AbortWithStatusJSON(500, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"name":     profile.Response.Players[0].PersonaName,
		"steamId":  steamID,
		"avatar":   profile.Response.Players[0].Avatar,
		"realName": profileData.RealName,
		"email":    profileData.Email,
		"phone":    profileData.Phone,
	})
}

func PatchProfileHandler(c *gin.Context) {
	steamID, exist := c.Get("steamID")
	if !exist {
		c.AbortWithStatusJSON(401, gin.H{"error": "authentication required"})
		return
	}

	var req updateUserProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "realName, email and phone are required"})
		return
	}

	req.RealName = strings.TrimSpace(req.RealName)
	req.Email = strings.TrimSpace(req.Email)
	req.Phone = strings.TrimSpace(req.Phone)

	if req.RealName == "" || req.Email == "" || req.Phone == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "realName, email and phone are required"})
		return
	}

	if !emailPattern.MatchString(req.Email) {
		c.AbortWithStatusJSON(400, gin.H{"error": "invalid email format"})
		return
	}

	if !taiwanMobilePattern.MatchString(req.Phone) {
		c.AbortWithStatusJSON(400, gin.H{"error": "invalid phone format"})
		return
	}

	collection := model.Db.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := collection.UpdateOne(
		ctx,
		bson.M{"SteamID": steamID.(string)},
		bson.M{
			"$set": bson.M{
				"realName": req.RealName,
				"email":    req.Email,
				"phone":    req.Phone,
			},
		},
	)
	if err != nil {
		log.Println("Error occurred while updating user profile:", err)
		c.AbortWithStatusJSON(500, gin.H{"error": "internal server error"})
		return
	}

	if result.MatchedCount == 0 {
		c.AbortWithStatusJSON(404, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"realName": req.RealName,
		"email":    req.Email,
		"phone":    req.Phone,
	})
}
