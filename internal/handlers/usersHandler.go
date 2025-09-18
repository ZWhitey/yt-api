package handlers

import (
	"context"
	"log"
	"net/http"
	"time"
	"yt-api/internal/model"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// User 表示從 users collection 取得的用戶資料
type User struct {
	SteamID string `bson:"SteamID" json:"SteamID"`
}

// UserResponse 表示 API 回傳的用戶資料格式
type UserResponse struct {
	SteamID string `json:"SteamID"`
	Name    string `json:"Name"`
}

// GetUsersHandler 處理 GET /api/v1/users 請求
func GetUsersHandler(c *gin.Context) {
	collection := model.Db.Collection("users")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Println("Error occurred while finding users:", err)
		c.AbortWithStatusJSON(500, gin.H{"error": "internal server error"})
		return
	}
	defer cursor.Close(ctx)

	var users []User
	if err = cursor.All(ctx, &users); err != nil {
		log.Println("Error occurred while decoding users:", err)
		c.AbortWithStatusJSON(500, gin.H{"error": "internal server error"})
		return
	}

	// 轉換為回應格式，Name 固定為 SteamID
	var response []UserResponse
	for _, user := range users {
		response = append(response, UserResponse{
			SteamID: user.SteamID,
			Name:    user.SteamID, // Name 固定回覆 SteamID
		})
	}

	c.JSON(http.StatusOK, response)
}
