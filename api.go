package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"
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

	router.GET("/api/v1/bot/status", getPriceHandler)

	router.Run(":8080")
}

func getPriceHandler(c *gin.Context) {
	var now = time.Now().Unix()
	if botStatusCache.Updated < now-60 {
		priceChan := make(chan int)
		go getPrice(priceChan)
		stockChan := make(chan int)
		go getStock(stockChan)
		orderChan := make(chan int)
		go getOrders(orderChan)
		botStatusCache.Price = <-priceChan
		botStatusCache.Stock = <-stockChan
		botStatusCache.Orders = <-orderChan
		botStatusCache.Updated = now
		fmt.Printf("Update cache to %+v\n", botStatusCache)
	}

	c.JSON(http.StatusOK, gin.H{
		"price":  botStatusCache.Price,
		"stock":  botStatusCache.Stock,
		"orders": botStatusCache.Orders,
	})
}

func getStock(resultChan chan<- int) {
	url := "https://steamcommunity.com/inventory/76561198047686623/440/2?l=english&count=1000"

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var inventory Inventory
	err = json.Unmarshal(body, &inventory)
	if err != nil {
		panic(err)
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
			if err == nil {
				resultChan <- price
			}
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
		log.Fatal(err)
		// 處理錯誤
	}
	print(count)
	resultChan <- int(count)
}
