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
	"go.mongodb.org/mongo-driver/mongo"
)

var mutex = &sync.Mutex{}

var botStatusCache BotStatus = BotStatus{
	Price:        0,
	Stock:        0,
	Orders:       0,
	Updated:      0,
	MarketPrice:  0,
	Transcations: 0,
}

func GetPriceHandler(c *gin.Context) {

	UpdateStatusCache()
	c.JSON(http.StatusOK, gin.H{
		"price":        botStatusCache.Price,
		"stock":        botStatusCache.Stock,
		"orders":       botStatusCache.Orders,
		"marketPrice":  botStatusCache.MarketPrice,
		"transactions": botStatusCache.Transcations,
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
	transactionChan := make(chan int)
	go getTransactions(transactionChan)

	botStatusCache.Price = <-priceChan
	botStatusCache.Stock = <-stockChan
	botStatusCache.Orders = <-orderChan
	botStatusCache.MarketPrice = <-marketPriceChan
	botStatusCache.Transcations = <-transactionChan

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

	condV2 := bson.M{
		"OrderStatus.Amt": bson.M{
			"$exists": true,
		},
	}

	countV2, errV2 := model.Db.Collection("orderv2").CountDocuments(context.TODO(), &condV2)

	if err != nil || errV2 != nil {
		log.Println("Error occurred while reading orders:", err)
		log.Println("Error occurred while reading orderv2:", errV2)
		resultChan <- botStatusCache.Orders
		return
	}
	resultChan <- int(count + countV2)
}

func getTransactions(resultChan chan<- int) {
	ctx := context.TODO()

	userTotal := 0
	transTotal := 0

	// Step 1: aggregate from "users"
	aggregation := mongo.Pipeline{
		bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$Transaction"}}}},
		bson.D{{Key: "$match", Value: bson.D{{Key: "Transaction.Traded", Value: true}}}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "total", Value: bson.D{{Key: "$sum", Value: "$Transaction.Count"}}},
		}}},
	}

	if userCur, err := model.Db.Collection("users").Aggregate(ctx, aggregation); err == nil {
		defer userCur.Close(ctx)
		if userCur.Next(ctx) {
			var res struct {
				Total int `bson:"total"`
			}
			if decodeErr := userCur.Decode(&res); decodeErr != nil {
				log.Println("Error decoding user aggregation:", decodeErr)
			} else {
				userTotal = res.Total
			}
		}
	} else {
		log.Println("Error aggregating users:", err)
	}

	// Step 2: sum from "transactions"
	cond := bson.M{"traded": true}
	if transCur, err := model.Db.Collection("transcations").Find(ctx, cond); err == nil {
		defer transCur.Close(ctx)
		for transCur.Next(ctx) {
			var t struct {
				Count int `bson:"Count"`
			}
			if decodeErr := transCur.Decode(&t); decodeErr != nil {
				log.Println("Error decoding transaction:", decodeErr)
				continue
			}
			transTotal += t.Count
		}
	} else {
		log.Println("Error querying transactions:", err)
	}

	// Step 3: return both totals summed
	resultChan <- userTotal + transTotal
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
