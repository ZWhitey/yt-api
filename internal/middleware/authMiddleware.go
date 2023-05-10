package middlewares

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	_ "github.com/joho/godotenv/autoload"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET"))

func AuthMiddleware(c *gin.Context) {
	session := sessions.Default(c)
	if session.Get("session") == nil {
		session.Delete("session")
		session.Save()
		c.AbortWithStatusJSON(401, gin.H{"error": "authentication required"})
		return
	}
	tokenString := session.Get("session").(string)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtKey, nil
	})
	if err != nil {
		log.Println("Parse jwt error:", err.Error())
		session.Delete("session")
		session.Save()
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		c.Set("steamID", claims["steamID"])
	} else {
		log.Println("Invalid jwt token", token.Raw)
		session.Delete("session")
		session.Save()
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	c.Next()
}
