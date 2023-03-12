package main

import (
	"net/http"
	"regexp"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"
)

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
	resultChan := make(chan int)
	go getPrice(resultChan)
	price := <-resultChan
	c.JSON(http.StatusOK, gin.H{
		"price":  price,
		"stock":  1000,
		"orders": 1000,
	})
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
