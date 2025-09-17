package middlewares

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	_ "github.com/joho/godotenv/autoload"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET"))

func AuthMiddleware(c *gin.Context) {
	// 直接從 cookie 讀取 JWT token
	tokenString, err := c.Cookie("session")
	if err != nil || tokenString == "" {
		c.AbortWithStatusJSON(401, gin.H{"error": "authentication required"})
		return
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtKey, nil
	})
	if err != nil {
		log.Println("Parse jwt error:", err.Error())
		// 清除無效的 cookie
		c.SetCookie("session", "", -1, "/", ".whitey.me", true, true)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		c.Set("steamID", claims["steamID"])
	} else {
		log.Println("Invalid jwt token", token.Raw)
		// 清除無效的 cookie
		c.SetCookie("session", "", -1, "/", ".whitey.me", true, true)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	c.Next()
}
