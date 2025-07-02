package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"yt-api/internal/model"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func validateSmilePay(purchamt, smseid, midSmilepay string) bool {
	merchantCode := os.Getenv("VERIFY_CODE")
	if merchantCode == "" {
		log.Println("VERIFY_CODE environment variable is not set")
		return false
	}
	A := fmt.Sprintf("%04s", merchantCode)

	amount, _ := strconv.Atoi(purchamt)
	B := fmt.Sprintf("%08d", amount)

	smseidLast4 := ""
	if len(smseid) >= 4 {
		smseidLast4 = smseid[len(smseid)-4:]
	} else {
		smseidLast4 = smseid
	}
	C := ""
	for _, char := range smseidLast4 {
		if char >= '0' && char <= '9' {
			C += string(char)
		} else {
			C += "9"
		}
	}

	C = fmt.Sprintf("%04s", C)

	D := A + B + C

	E := 0
	for i, char := range D {
		if (i+1)%2 == 0 {
			digit, _ := strconv.Atoi(string(char))
			E += digit
		}
	}
	E *= 3

	F := 0
	for i, char := range D {
		if (i+1)%2 == 1 {
			digit, _ := strconv.Atoi(string(char))
			F += digit
		}
	}
	F *= 9

	calculatedCode := E + F
	receivedCode, _ := strconv.Atoi(midSmilepay)
	return calculatedCode == receivedCode
}

func PaymentCallbackHandler(c *gin.Context) {
	ROTURL_STATUS := os.Getenv("ROTURL_STATUS")
	if ROTURL_STATUS == "" {
		log.Println("ROTURL_STATUS environment variable is not set")
		c.Data(http.StatusInternalServerError, "text/html; charset=utf-8", []byte("Internal Server Error"))
		return
	}

	err := c.Request.ParseForm()
	if err != nil {
		log.Printf("Error parsing form data: %v\n", err)
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte("Invalid request format"))
		return
	}

	dataId := c.PostForm("Data_id")
	processDate := c.PostForm("Process_date")
	processTime := c.PostForm("Process_time")
	purchamt := c.PostForm("Purchamt") // purchamt is the amount of the purchase
	amount := c.PostForm("Amount")     // Amount is the amount received from the payment gateway
	midSmilepay := c.PostForm("Mid_smilepay")
	smseid := c.PostForm("Smseid")

	validationResult := validateSmilePay(purchamt, smseid, midSmilepay)
	if !validationResult {
		log.Printf("Validation failed, not a valid SmilePay transaction: dataId=%s purchamt=%s, smseid=%s, midSmilepay=%s, amount=%s\n", dataId, purchamt, smseid, midSmilepay, amount)
		c.Data(http.StatusForbidden, "text/html; charset=utf-8", []byte("Invalid SmilePay transaction"))
		return
	}

	pcthAmt, err := strconv.Atoi(purchamt)
	if err != nil {
		log.Printf("Error converting purchamt to integer: %v\n", err)
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte("Invalid purchamt format"))
		return
	}

	amt, err := strconv.Atoi(amount)
	if err != nil {
		log.Printf("Error converting amount to integer: %v\n", err)
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte("Invalid amount format"))
		return
	}
	if amt != pcthAmt {
		log.Printf("Amount mismatch: purchamt=%d, amount=%d\n", pcthAmt, amt)
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte("Amount mismatch"))
		return
	}
	collection := model.Db.Collection("orderv2")

	filter := bson.M{
		"OrderStatus.Data_id": dataId,
	}

	update := bson.M{
		"$set": bson.M{
			"OrderStatus.Process_date": processDate,
			"OrderStatus.Process_time": processTime,
			"OrderStatus.Amt":          amt,
		},
	}
	result := collection.FindOneAndUpdate(context.TODO(), filter, update)
	if result.Err() != nil {
		log.Printf("Error updating order: %v\n", result.Err())
		c.Data(http.StatusInternalServerError, "text/html; charset=utf-8", []byte("Internal Server Error"))
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("<Roturlstatus>"+ROTURL_STATUS+"</Roturlstatus>"))
}
