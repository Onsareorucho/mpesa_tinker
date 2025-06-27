package mpesa

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
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

type RegisterUrlRequest struct {
	ResponseType    string `json:"ResponseType"`
	ConfirmationURL string `json:"ConfirmationURL"`
	ValidationURL   string `json:"ValidationURL"`
}

type RegisterUrlResponse struct {
	OriginatorCoversationID string `json:"OriginatorCoversationID"`
	ResponseCode            string `json:"ResponseCode"`
	ResponseDescription     string `json:"ResponseDescription"`
}

// {
//    "ShortCode": "601426",
//    "ResponseType":"[Cancelled/Completed]",
//    "ConfirmationURL":"[confirmation URL]",
//    "ValidationURL":"[validation URL]"
// }

func RegisterUrlHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req RegisterUrlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ResponseType == "" ||
		req.ConfirmationURL == "" || req.ValidationURL == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tm := &TokenManager{}
	token, err := tm.GetToken()
	if err != nil {
		log.Println("❌ Failed to get token:", err)
		http.Error(w, "Failed to get token", http.StatusInternalServerError)
		return
	}

	RegUrlResp, err := InitiateRegisterUrl(token, req.ResponseType, req.ConfirmationURL, req.ValidationURL)
	if err != nil {
		log.Println("❌ Register URL error:", err)
		http.Error(w, "Failed to register the URL", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(RegUrlResp)
}

func InitiateRegisterUrl(token, responseType, confirmationURL, validationURL string) (*RegisterUrlResponse, error) {
    payload := map[string]string{}

    body, _ := json.Marshal(payload)
    req, _ := http.NewRequest("POST",
        "https://sandbox.safaricom.co.ke/mpesa/c2b/v1/registerurl",
        bytes.NewBuffer(body))

    req.Header.Set("Authorization", "Bearer "+token)
    req.Header.Set("Content-Type", "application/json")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    respBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("read response: %v", err)
    }
    log.Println("📦 Daraja Register URL response:", string(respBody))

    var registerResp RegisterUrlResponse
    if err := json.Unmarshal(respBody, &registerResp); err != nil {
        return nil, fmt.Errorf("decode response: %v", err)
    }

    return &registerResp, nil
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
		"CallBackURL":       os.Getenv("MPESA_CALLBACK_URL"),
		"AccountReference":  "YOURFAVORITEDEVELOPER",
		"TransactionDesc":   "Payment of X",
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "https://sandbox.safaricom.co.ke/mpesa/stkpush/v1/processrequest", bytes.NewBuffer(body))
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")
	log.Println("Callback URL:", os.Getenv("MPESA_CALLBACK_URL"))
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
