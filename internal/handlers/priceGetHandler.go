package handlers

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"time"

	"yt-api/internal/model"
	. "yt-api/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"
	"go.mongodb.org/mongo-driver/bson"
)

var mutex = &sync.Mutex{}

var botStatusCache BotStatus = BotStatus{
	Price:   0,
	Stock:   0,
	Orders:  0,
	Updated: 0,
	MarketPrice: 0,
}

func GetPriceHandler(c *gin.Context) {

	UpdateStatusCache()
	c.JSON(http.StatusOK, gin.H{
		"price":  botStatusCache.Price,
		"stock":  botStatusCache.Stock,
		"orders": botStatusCache.Orders,
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

	body, err := ioutil.ReadAll(resp.Body)
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