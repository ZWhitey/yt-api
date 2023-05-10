package handlers

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	_ "github.com/joho/godotenv/autoload"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET"))

func AuthHandler(c *gin.Context) {
	session := sessions.Default(c)
	ns := c.Query("openid.ns")
	claimedID := c.Query("openid.claimed_id")
	identity := c.Query("openid.identity")
	returnTo := c.Query("openid.return_to")
	nonce := c.Query("openid.response_nonce")
	assocHandle := c.Query("openid.assoc_handle")
	signedParams := c.Query("openid.signed")
	signature := c.Query("openid.sig")

	params := url.Values{}
	params.Set("openid.ns", ns)
	params.Set("openid.mode", "check_authentication")
	params.Set("openid.op_endpoint", "https://steamcommunity.com/openid/login")
	params.Set("openid.claimed_id", claimedID)
	params.Set("openid.identity", identity)
	params.Set("openid.return_to", returnTo)
	params.Set("openid.response_nonce", nonce)
	params.Set("openid.assoc_handle", assocHandle)
	params.Set("openid.signed", signedParams)
	params.Set("openid.sig", signature)

	// 發送 POST 請求
	resp, err := http.PostForm("https://steamcommunity.com/openid/login", params)
	if err != nil {
		log.Println("Check Steam openid error:", err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// 讀取回應內容
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Parse Steam openid response error:", err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// check is_valid:true
	if !strings.Contains(string(body), "is_valid:true") {
		log.Println("Invalid openid data:", string(body))
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// 解析 Steam ID
	steamID := strings.TrimPrefix(identity, "https://steamcommunity.com/openid/id/")
	if steamID == "" {
		log.Println("Cannot find steamID:", identity)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["steamID"] = steamID
	claims["exp"] = time.Now().Add(time.Hour * 168).Unix() // 1 week
	sessionKey, err := token.SignedString(jwtKey)
	if err != nil {
		session.Delete("session")
		session.Save()
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	session.Set("session", sessionKey)
	session.Save()

	c.Redirect(http.StatusFound, "https://tf2key.whitey.me/")
}
