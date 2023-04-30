package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"
	"github.com/golang-jwt/jwt/v5"
	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Inventory struct {
	Assets []struct {
		Appid      int    `json:"appid"`
		Contextid  string `json:"contextid"`
		Assetid    string `json:"assetid"`
		Classid    string `json:"classid"`
		Instanceid string `json:"instanceid"`
		Amount     string `json:"amount"`
	} `json:"assets"`
	Descriptions []struct {
		Appid            int    `json:"appid"`
		Classid          string `json:"classid"`
		Instanceid       string `json:"instanceid"`
		Currency         int    `json:"currency"`
		Background_color string `json:"background_color"`
		Icon_url         string `json:"icon_url"`
		Icon_url_large   string `json:"icon_url_large"`
		Descriptions     []struct {
			Value string `json:"value"`
			Color string `json:"color,omitempty"`
		} `json:"descriptions"`
		Tradable int `json:"tradable"`
		Actions  []struct {
			Link string `json:"link"`
			Name string `json:"name"`
		} `json:"actions"`
		Name                          string `json:"name"`
		Name_color                    string `json:"name_color"`
		Type                          string `json:"type"`
		Market_name                   string `json:"market_name"`
		Market_hash_name              string `json:"market_hash_name"`
		Commodity                     int    `json:"commodity"`
		Market_tradable_restriction   int    `json:"market_tradable_restriction"`
		Market_marketable_restriction int    `json:"market_marketable_restriction"`
		Marketable                    int    `json:"marketable"`
		Tags                          []struct {
			Category                string `json:"category"`
			Internal_name           string `json:"internal_name"`
			Localized_category_name string `json:"localized_category_name"`
			Localized_tag_name      string `json:"localized_tag_name"`
			Color                   string `json:"color"`
		} `json:"tags"`
	} `json:"descriptions"`
	More_items            int    `json:"more_items"`
	Last_assetid          string `json:"last_assetid"`
	Total_inventory_count int    `json:"total_inventory_count"`
	Success               int    `json:"success"`
	Rwgrsn                int    `json:"rwgrsn"`
}

type BotStatus struct {
	Price   int
	Stock   int
	Orders  int
	Updated int64
}

var botStatusCache BotStatus = BotStatus{
	Price:   0,
	Stock:   0,
	Orders:  0,
	Updated: 0,
}

var mgoClient *mongo.Client
var db *mongo.Database

var jwtKey = []byte(os.Getenv("JWT_SECRET"))

func main() {

	var username = os.Getenv("MONGO_USERNAME")
	var password = os.Getenv("MONGO_PASSWORD")
	connectionString := fmt.Sprintf("mongodb+srv://%s:%s@cluster0.bkfvo.mongodb.net/?retryWrites=true&w=majority", username, password)
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().
		ApplyURI(connectionString).
		SetServerAPIOptions(serverAPIOptions)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mgoClient, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = mgoClient.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	db = mgoClient.Database("heroku_x68nnv7z")

	router := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://127.0.0.1:5173", "https://tf2key.whitey.me"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE"}
	router.Use(cors.New(config))

	authMiddleware := func() gin.HandlerFunc {
		return func(c *gin.Context) {
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
	}

	store := cookie.NewStore([]byte("key"))
	store.Options(sessions.Options{
		Domain: ".whitey.me",
	})
	router.Use(sessions.Sessions("session", store))

	router.GET("/api/v1/bot/status", getPriceHandler)
	router.GET("/auth", authHandler)
	router.GET("/api/v1/orders", authMiddleware(), orderHandler)

	router.Run(":8080")
	updateStatusCache()
}

func orderHandler(c *gin.Context) {
	steamID, exist := c.Get("steamID")
	if exist {
		c.JSON(http.StatusOK, gin.H{
			"id": steamID,
		})
		return
	}
	c.AbortWithStatusJSON(401, gin.H{"error": "authentication required"})
}

func authHandler(c *gin.Context) {
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

var mutex = &sync.Mutex{}

func getPriceHandler(c *gin.Context) {

	updateStatusCache()
	c.JSON(http.StatusOK, gin.H{
		"price":  botStatusCache.Price,
		"stock":  botStatusCache.Stock,
		"orders": botStatusCache.Orders,
	})
}

func updateStatusCache() {
	now := time.Now().Unix()
	mutex.Lock()
	defer mutex.Unlock()
	if botStatusCache.Updated+300 > now {
		return
	}
	log.Println("Start update cache")

	priceChan := make(chan int)
	go getPrice(priceChan)
	stockChan := make(chan int)
	go getStock(stockChan)
	orderChan := make(chan int)
	go getOrders(orderChan)
	botStatusCache.Price = <-priceChan
	botStatusCache.Stock = <-stockChan
	botStatusCache.Orders = <-orderChan
	botStatusCache.Updated = time.Now().Unix()
	log.Printf("Update cache to %+v\n", botStatusCache)
}

func getStock(resultChan chan<- int) {
	resultChan <- botStatusCache.Stock

	url := "https://steamcommunity.com/inventory/76561198047686623/440/2?l=english&count=1000"

	resp, err := http.Get(url)
	if err != nil {
		log.Println("Error occurred while sending request:", err)
		resultChan <- botStatusCache.Stock // indicate error to the caller
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error occurred while reading response body:", err)
		resultChan <- botStatusCache.Stock // indicate error to the caller
		return
	}

	var inventory Inventory
	err = json.Unmarshal(body, &inventory)
	if err != nil {
		log.Println("Error occurred while unmarshaling JSON:", err)
		resultChan <- botStatusCache.Stock // indicate error to the caller
		return
	}

	count := 0
	for _, item := range inventory.Assets {
		if item.Classid == "101785959" {
			count += 1
		}

	}
	resultChan <- count
}

func getPrice(resultChan chan<- int) {
	c := colly.NewCollector()

	// 抓類別Class 名稱
	c.OnHTML(".profile_summary", func(e *colly.HTMLElement) {
		re := regexp.MustCompile(`(\d+)\s*元`)
		matches := re.FindStringSubmatch(e.Text)
		if len(matches) > 1 {
			price, err := strconv.Atoi(matches[1])
			if err != nil {
				log.Println("Error occurred while parsing price:", err)
				resultChan <- botStatusCache.Price
				return
			}
			resultChan <- price
		}

	})

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.75 Safari/537.36")
	})

	c.Visit("https://steamcommunity.com/id/Whitey_Keybot/")
}

func getOrders(resultChan chan<- int) {
	var collection = db.Collection("orders")
	cond := bson.M{
		"OrderStatus.TradeStatus": "1",
	}
	count, err := collection.CountDocuments(context.TODO(), &cond)
	if err != nil {
		log.Println("Error occurred while reading orders:", err)
		resultChan <- botStatusCache.Orders
		return
	}
	resultChan <- int(count)
}
