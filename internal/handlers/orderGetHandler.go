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

type order struct {
	SteamID     string
	Price       int
	Count       int
	OrderStatus struct {
		TradeNo     string
		PayInfo     string
		PaymentType string
		TradeStatus string
		ExpireDate  string
	}
}

func GetOrderHandler(c *gin.Context) {
	steamID, exist := c.Get("steamID")
	if !exist {
		c.AbortWithStatusJSON(401, gin.H{"error": "authentication required"})
		return
	}
	id := c.Query("id")
	if id != "" && (steamID == "76561198041578278" || steamID == "76561198047686623") {
		log.Println("Admin override", steamID, id)
		steamID = id
	}
	collection := model.Db.Collection("orders")
	query := bson.M{"SteamID": steamID}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cursor, err := collection.Find(ctx, &query)
	if err != nil {
		log.Println("Error occurred while finding orders:", err)
		c.AbortWithStatusJSON(500, gin.H{"error": "internal server error"})
		return
	}

	defer func() {
		if err = cursor.Close(ctx); err != nil {
			log.Fatal(err)
		}
	}()
	var orders []order
	if err := cursor.All(ctx, &orders); err != nil {
		log.Println("Error occurred while reading orders:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"orders": orders,
	})
}
