package main

import (
	"os"
	. "yt-api/internal/handlers"
	. "yt-api/internal/middleware"
	"yt-api/internal/model"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	// 初始化 Redis 連接
	model.InitRedis()
	defer model.CloseRedis()

	port := "8080"

	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}

	router := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://local.whitey.me:5173", "https://local.whitey.me:5173", "https://tf2key.whitey.me"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE"}
	config.AllowCredentials = true
	router.Use(cors.New(config))

	router.GET("/api/v1/bot/status", GetPriceHandler)
	router.GET("/auth", AuthHandler)
	router.GET("/api/v1/orders", AuthMiddleware, GetOrderHandler)
	router.GET("/api/v2/orders", AuthMiddleware, GetOrderV2Handler)
	router.GET("api/v1/user", AuthMiddleware, GetProfileHandler)
	router.POST("/api/v1/payment/cb", PaymentCallbackHandler)

	router.Run(":" + port)
	UpdateStatusCache()
}
