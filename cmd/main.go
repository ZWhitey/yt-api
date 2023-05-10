package main

import (
	"net/http"
	"os"
	. "yt-api/internal/handlers"
	. "yt-api/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
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

	store := cookie.NewStore([]byte("key"))
	store.Options(sessions.Options{
		Domain:   ".whitey.me",
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
		HttpOnly: true,
	})
	router.Use(sessions.Sessions("session", store))

	router.GET("/api/v1/bot/status", GetPriceHandler)
	router.GET("/auth", AuthHandler)
	router.GET("/api/v1/orders", AuthMiddleware, OrderHandler)

	router.Run(":" + port)
	UpdateStatusCache()
}
