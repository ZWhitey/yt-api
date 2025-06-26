package handlers

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"yt-api/internal/model"
	. "yt-api/internal/types"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

var mutex = &sync.Mutex{}

var botStatusCache BotStatus = BotStatus{
	Price:       0,
	Stock:       0,
	Orders:      0,
	Updated:     0,
	MarketPrice: 0,
}

func GetPriceHandler(c *gin.Context) {

	UpdateStatusCache()
	c.JSON(http.StatusOK, gin.H{
		"price":       botStatusCache.Price,
		"stock":       botStatusCache.Stock,
		"orders":      botStatusCache.Orders,
		"marketPrice": botStatusCache.MarketPrice,
	})
}

func UpdateStatusCache() {
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
	marketPriceChan := make(chan int)
	go getMarketPrice(marketPriceChan)
	botStatusCache.Price = <-priceChan
	botStatusCache.Stock = <-stockChan
	botStatusCache.Orders = <-orderChan
	botStatusCache.MarketPrice = <-marketPriceChan
	botStatusCache.Updated = time.Now().Unix()
	log.Printf("Update cache to %+v\n", botStatusCache)
}

func getStock(resultChan chan<- int) {
	ctx := context.Background()

	// 從 Redis 獲取庫存數據
	stockStr, err := model.RedisClient.Get(ctx, "REDIS_STOCK").Result()
	if err != nil {
		log.Printf("Error getting stock from Redis: %v", err)
		// 如果 Redis 獲取失敗，返回快取值
		resultChan <- botStatusCache.Stock
		return
	}

	stock, err := strconv.Atoi(stockStr)
	if err != nil {
		log.Printf("Error parsing stock value from Redis: %v", err)
		resultChan <- botStatusCache.Stock
		return
	}

	resultChan <- stock
}

func getPrice(resultChan chan<- int) {
	ctx := context.Background()

	// 從 Redis 獲取價格數據
	priceStr, err := model.RedisClient.Get(ctx, "REDIS_PRICE").Result()
	if err != nil {
		log.Printf("Error getting price from Redis: %v", err)
		// 如果 Redis 獲取失敗，返回快取值
		resultChan <- botStatusCache.Price
		return
	}

	price, err := strconv.Atoi(priceStr)
	if err != nil {
		log.Printf("Error parsing price value from Redis: %v", err)
		resultChan <- botStatusCache.Price
		return
	}

	resultChan <- price
}

func getOrders(resultChan chan<- int) {
	var collection = model.Db.Collection("orders")
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

func getMarketPrice(resultChan chan<- int) {
	url := "https://steamcommunity.com/market/itemordershistogram?country=TW&language=tchinese&currency=30&item_nameid=1&two_factor=0"

	resp, err := http.Get(url)
	if err != nil {
		log.Println("Error occurred while sending request:", err)
		resultChan <- botStatusCache.MarketPrice // indicate error to the caller
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error occurred while reading response body:", err)
		resultChan <- botStatusCache.MarketPrice // indicate error to the caller
		return
	}

	var item MarketItem
	err = json.Unmarshal(body, &item)
	if err != nil {
		log.Println("Error occurred while unmarshaling JSON:", err)
		resultChan <- botStatusCache.MarketPrice // indicate error to the caller
		return
	}

	price, err := strconv.Atoi(item.LowestSellOrder)
	if err != nil {
		log.Println("Error:", err)
		resultChan <- botStatusCache.MarketPrice // indicate error to the caller
		return
	}

	resultChan <- price
}
