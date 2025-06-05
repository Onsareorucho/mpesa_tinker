package mpesa

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"
)

type STKResponse struct {
	ResponseCode        string `json:"ResponseCode"`
	CustomerMessage     string `json:"CustomerMessage"`
	MerchantRequestID   string `json:"MerchantRequestID"`
	CheckoutRequestID   string `json:"CheckoutRequestID"`
	ResponseDescription string `json:"ResponseDescription"`
}

func InitiateSTKPush(token, phone, amount string) (*STKResponse, error) {
	timestamp := time.Now().Format("20060102150405")
	shortcode := os.Getenv("MPESA_SHORTCODE")
	passkey := os.Getenv("MPESA_PASSKEY")
	password := base64.StdEncoding.EncodeToString([]byte(shortcode + passkey + timestamp))

	amountInt, err := strconv.Atoi(amount)
	if err != nil {
		return nil, err
	}

	phoneInt, err := strconv.Atoi(phone)
	if err != nil {
		return nil, err
	}

	payload := map[string]interface{}{
		"BusinessShortCode": shortcode,
		"Password":          password,
		"Timestamp":         timestamp,
		"TransactionType":   "CustomerPayBillOnline",
		"Amount":            amountInt,
		"PartyA":            phoneInt,
		"PartyB":            174379,
		"PhoneNumber":       phoneInt,
		"CallBackURL":       "https://mydomain.com/path",
		"AccountReference":  "CompanyXLTD",
		"TransactionDesc":   "Payment of X",
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "https://sandbox.safaricom.co.ke/mpesa/stkpush/v1/processrequest", bytes.NewBuffer(body))
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	// return client.Do(req)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var stkResp STKResponse
	err = json.NewDecoder(resp.Body).Decode(&stkResp)
	if err != nil {
		return nil, err
	}

	return &stkResp, nil
}
