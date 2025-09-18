package handlers

import (
	"context"
	"log"
	"net/http"
	"time"
	"yt-api/internal/model"
	"yt-api/internal/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// User 表示從 users collection 取得的用戶資料
type User struct {
	SteamID string `bson:"SteamID" json:"SteamID"`
}

// UserResponse 表示 API 回傳的用戶資料格式
type UserResponse struct {
	SteamID string `json:"steamId"`
	Name    string `json:"name"`
}

// UserDetail 表示用戶詳細資料的回應格式
type UserDetail struct {
	SteamID         string `json:"steamId"`
	Name            string `json:"name"`
	AvatarURL       string `json:"avatarUrl"`
	UnclaimedCount  int    `json:"unclaimedCount"`
	ClaimedCount    int    `json:"claimedCount"`
	CompletedOrders int    `json:"completedOrders"`
	ActiveOrders    int    `json:"activeOrders"`
}

// OrderV2 表示從 orderv2 collection 取得的訂單資料
type OrderV2 struct {
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

// Transaction 表示從 transactions collection 取得的交易資料
type Transaction struct {
	ID        string    `bson:"_id,omitempty" json:"_id,omitempty"`
	SteamID   string    `bson:"steamID" json:"steamID"`
	TradeID   string    `bson:"tradeId" json:"tradeId"`
	Count     int       `bson:"Count" json:"Count"`
	Traded    bool      `bson:"traded" json:"traded"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`
	V         int       `bson:"__v" json:"__v"`
}

// TransactionResponse 表示交易記錄的回應格式
type TransactionResponse struct {
	ID        string `json:"id"`
	Quantity  int    `json:"quantity"`
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

// GetUsersHandler 處理 GET /api/v1/users 請求
func GetUsersHandler(c *gin.Context) {
	steamID, exist := c.Get("steamID")
	if !exist {
		c.AbortWithStatusJSON(401, gin.H{"error": "authentication required"})
		return
	}

	if !ADMIN_STEAM_ID_SET[steamID.(string)] {
		c.AbortWithStatusJSON(403, gin.H{"error": "forbidden"})
		return
	}

	collection := model.Db.Collection("users")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Println("Error occurred while finding users:", err)
		c.AbortWithStatusJSON(500, gin.H{"error": "internal server error"})
		return
	}
	defer cursor.Close(ctx)

	var users []User
	if err = cursor.All(ctx, &users); err != nil {
		log.Println("Error occurred while decoding users:", err)
		c.AbortWithStatusJSON(500, gin.H{"error": "internal server error"})
		return
	}

	// 轉換為回應格式，Name 固定為 SteamID
	var response []UserResponse
	for _, user := range users {
		response = append(response, UserResponse{
			SteamID: user.SteamID,
			Name:    user.SteamID, // Name 固定回覆 SteamID
		})
	}

	c.JSON(http.StatusOK, response)
}

// getUserOrderStats 根據 SteamID 獲取用戶的訂單統計資料
func getUserOrderStats(steamID string) (completedOrders, activeOrders, payedAmount int, err error) {
	collection := model.Db.Collection("orderv2")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 查詢該用戶的所有訂單
	cursor, err := collection.Find(ctx, bson.M{"SteamID": steamID})
	if err != nil {
		log.Printf("Error finding orders for SteamID %s: %v", steamID, err)
		return 0, 0, 0, err
	}
	defer cursor.Close(ctx)

	var orders []OrderV2
	if err = cursor.All(ctx, &orders); err != nil {
		log.Printf("Error decoding orders for SteamID %s: %v", steamID, err)
		return 0, 0, 0, err
	}

	// 統計訂單狀態
	for _, order := range orders {
		// 判斷訂單是否已完成 (有 ProcessDate 和 ProcessTime 表示已付款完成)
		payEndDate, err := time.Parse("2006/01/02 15:04:05", order.OrderStatus.PayEndDate)
		if err != nil {
			log.Println("Error parsing PayEndDate:", err)
			payEndDate = time.Time{}
		}
		if order.OrderStatus.Amt == order.OrderStatus.Amount {
			completedOrders++
			payedAmount += order.Count
		} else if payEndDate.After(time.Now()) {
			activeOrders++
		}
	}

	return completedOrders, activeOrders, payedAmount, nil
}

func getUserTradedAmount(steamID string) (tradedAmount int, err error) {
	collection := model.Db.Collection("transcations")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cursor, err := collection.Find(ctx, bson.M{"steamID": steamID, "traded": true})
	if err != nil {
		log.Printf("Error counting traded transactions for SteamID %s: %v", steamID, err)
		return 0, err
	}

	var transactions []Transaction
	if err = cursor.All(ctx, &transactions); err != nil {
		log.Printf("Error decoding transactions for SteamID %s: %v", steamID, err)
		return 0, err
	}
	count := 0
	for _, txn := range transactions {
		count += txn.Count
	}

	return count, nil
}

// getUserTransactions 根據 SteamID 獲取用戶的交易記錄
func getUserTransactions(steamID string) ([]TransactionResponse, error) {
	collection := model.Db.Collection("transcations")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 查詢該用戶的所有交易記錄，按時間倒序排列
	cursor, err := collection.Find(ctx, bson.M{"steamID": steamID}, &options.FindOptions{
		Sort: bson.D{{Key: "_id", Value: -1}},
	})
	if err != nil {
		log.Printf("Error finding transactions for SteamID %s: %v", steamID, err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var transactions []Transaction
	if err = cursor.All(ctx, &transactions); err != nil {
		log.Printf("Error decoding transactions for SteamID %s: %v", steamID, err)
		return nil, err
	}

	// 轉換為回應格式
	var response []TransactionResponse
	for _, txn := range transactions {
		status := "失敗"
		if txn.Traded {
			status = "成功"
		}

		// 格式化時間為 "YYYY-MM-DD HH:mm" 格式
		timestamp := txn.CreatedAt.Format("2006-01-02 15:04")

		response = append(response, TransactionResponse{
			ID:        txn.TradeID, // 使用 tradeId 作為交易 ID
			Quantity:  txn.Count,   // 使用 Count 作為數量
			Status:    status,
			Timestamp: timestamp,
		})
	}

	return response, nil
}

// GetUserDetailHandler 處理 GET /api/v1/users/{id} 請求
func GetUserDetailHandler(c *gin.Context) {
	steamID, exist := c.Get("steamID")
	if !exist {
		c.AbortWithStatusJSON(401, gin.H{"error": "authentication required"})
		return
	}

	if !ADMIN_STEAM_ID_SET[steamID.(string)] {
		c.AbortWithStatusJSON(403, gin.H{"error": "forbidden"})
		return
	}

	targetId := c.Param("id")
	if targetId == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "steam id is required"})
		return
	}

	// 獲取訂單統計資料
	completedOrders, activeOrders, payAmount, err := getUserOrderStats(targetId)
	if err != nil {
		log.Printf("Error getting order stats for SteamID %s: %v", targetId, err)
		c.AbortWithStatusJSON(500, gin.H{"error": "internal server error"})
		return
	}

	// 獲取交易成功的次數
	tradedAmount, err := getUserTradedAmount(targetId)
	if err != nil {
		log.Printf("Error getting traded amount for SteamID %s: %v", targetId, err)
		c.AbortWithStatusJSON(500, gin.H{"error": "internal server error"})
		return
	}

	// 獲取 Steam 用戶資料
	profile, err := utils.GetProfileFromSteam(targetId)
	var name, avatarURL string
	if err != nil {
		log.Printf("Error getting Steam profile for SteamID %s: %v", targetId, err)
		// 如果無法獲取 Steam 資料，使用預設值
		name = targetId
		avatarURL = ""
	} else if len(profile.Response.Players) > 0 {
		name = profile.Response.Players[0].PersonaName
		avatarURL = profile.Response.Players[0].AvatarFull
	}

	userDetail := UserDetail{
		SteamID:         targetId,
		Name:            name,
		AvatarURL:       avatarURL,
		UnclaimedCount:  payAmount - tradedAmount,
		ClaimedCount:    tradedAmount,
		CompletedOrders: completedOrders,
		ActiveOrders:    activeOrders,
	}

	c.JSON(http.StatusOK, userDetail)
}

// GetUserTransactionsHandler 處理 GET /api/v1/users/{id}/transactions 請求
func GetUserTransactionsHandler(c *gin.Context) {
	steamID, exist := c.Get("steamID")
	if !exist {
		c.AbortWithStatusJSON(401, gin.H{"error": "authentication required"})
		return
	}

	if !ADMIN_STEAM_ID_SET[steamID.(string)] {
		c.AbortWithStatusJSON(403, gin.H{"error": "forbidden"})
		return
	}

	targetId := c.Param("id")
	if targetId == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "steam id is required"})
		return
	}

	// 獲取用戶的交易記錄
	transactions, err := getUserTransactions(targetId)
	if err != nil {
		log.Printf("Error getting transactions for SteamID %s: %v", targetId, err)
		c.AbortWithStatusJSON(500, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, transactions)
}
