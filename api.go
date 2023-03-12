package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"
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

func main() {
	router := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://127.0.0.1:5173"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE"}
	router.Use(cors.New(config))

	router.GET("/api/v1/bot/status", getPriceHandler)

	router.Run(":8080")
}

func getPriceHandler(c *gin.Context) {
	priceChan := make(chan int)
	go getPrice(priceChan)
	stockChan := make(chan int)
	go getStock(stockChan)
	price := <-priceChan
	stock := <-stockChan
	c.JSON(http.StatusOK, gin.H{
		"price":  price,
		"stock":  stock,
		"orders": 1000,
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
