package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"yt-api/internal/model"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

type orderv2 struct {
	SteamID     string `bson:"SteamID" json:"SteamID"`
	Price       int    `bson:"Price" json:"Price"`
	Count       int    `bson:"Count" json:"Count"`
	OrderStatus struct {
		SmilePayNO string `bson:"SmilePayNO" json:"SmilePayNO"`
		DataID     string `bson:"Data_id" json:"Data_id"`
		Amount     int    `bson:"Amount" json:"Amount"`
		PayEndDate string `bson:"PayEndDate" json:"PayEndDate"`
		PayMethod  string `bson:"PayMethod" json:"PayMethod"`
		AtmBankNo  string `bson:"AtmBankNo" json:"AtmBankNo"`
		AtmNo      string `bson:"AtmNo" json:"AtmNo"`
		IbonNo     string `bson:"IbonNo" json:"IbonNo"`
		FamiNO     string `bson:"FamiNO" json:"FamiNO"`
		// Callback Data
		ProcessDate string `bson:"Process_date" json:"Process_date"`
		ProcessTime string `bson:"Process_time" json:"Process_time"`
		Amt         int    `bson:"Amt" json:"Amt"`
	} `bson:"OrderStatus" json:"OrderStatus"`
}

type orderV2Response struct {
	SteamID   string `json:"SteamID"`
	Price     int    `json:"Price"`
	Count     int    `json:"Count"`
	Amount    int    `json:"Amount"`
	OrderId   string `json:"OrderId"`
	OrderDate string `json:"OrderDate"`
	PayDate   string `json:"PayDate,omitempty"`
	PayMethod string `json:"PayMethod"`
	Status    string `json:"Status"`
}

type Status string

const (
	StatusUnpaid  Status = "Unpaid"
	StatusPaid    Status = "Paid"
	StatusExpired Status = "Expired"
)

var ADMIN_STEAM_ID_SET = map[string]bool{
	"76561198041578278": true,
	"76561198047686623": true,
}

// parseDateTime 將 YYYYMMDDHHmmss 格式轉換為 YYYY/MM/DD HH:mm:ss 格式
func parseDateTime(dateTimeStr string) string {
	if len(dateTimeStr) != 14 {
		return dateTimeStr // 如果格式不正確，返回原字串
	}

	// 解析時間
	t, err := time.Parse("20060102150405", dateTimeStr)
	if err != nil {
		return dateTimeStr // 如果解析失敗，返回原字串
	}

	// 格式化為 YYYY/MM/DD HH:mm:ss
	return t.Format("2006/01/02 15:04:05")
}

func GetOrderV2Handler(c *gin.Context) {
	steamID, exist := c.Get("steamID")
	if !exist {
		c.AbortWithStatusJSON(401, gin.H{"error": "authentication required"})
		return
	}

	if !ADMIN_STEAM_ID_SET[steamID.(string)] {
		c.AbortWithStatusJSON(403, gin.H{"error": "forbidden"})
		return
	}

	collection := model.Db.Collection("orderv2")
	query := bson.M{}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cursor, err := collection.Find(ctx, &query)
	if err != nil {
		log.Println("Error occurred while finding orders:", err)
		c.AbortWithStatusJSON(500, gin.H{"error": "internal server error"})
		return
	}

	defer func() {
		if err = cursor.Close(ctx); err != nil {
			log.Fatal(err)
		}
	}()
	var orders []orderv2
	if err := cursor.All(ctx, &orders); err != nil {
		log.Println("Error occurred while reading orders:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	var responseOrders []orderV2Response
	for _, order := range orders {
		status := StatusUnpaid
		var payDate string

		payEndDate, err := time.Parse("2006/01/02 15:04:05", order.OrderStatus.PayEndDate)
		if err != nil {
			log.Println("Error parsing PayEndDate:", err)
			payEndDate = time.Time{} // 設定為零值，表示無效日期
		}

		if order.OrderStatus.Amt == order.OrderStatus.Amount {
			status = StatusPaid
			payDate = order.OrderStatus.ProcessDate + " " + order.OrderStatus.ProcessTime
		} else if payEndDate.Before(time.Now()) {
			status = StatusExpired
		}

		responseOrders = append(responseOrders, orderV2Response{
			SteamID:   order.SteamID,
			Price:     order.Price,
			Count:     order.Count,
			Amount:    order.OrderStatus.Amount,
			OrderId:   order.OrderStatus.DataID,
			OrderDate: parseDateTime(order.OrderStatus.DataID),
			PayDate:   payDate,
			PayMethod: order.OrderStatus.PayMethod,
			Status:    string(status),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"orders": responseOrders,
	})
}
