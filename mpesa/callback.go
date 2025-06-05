package mpesa

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"goRestAPI/db"
)

type MetadataItem struct {
	Name  string      `json:"Name"`
	Value interface{} `json:"Value"`
}

// CallbackRequest defines expected M-Pesa callback fields
type CallbackRequest struct {
	Body struct {
		StkCallback struct {
			MerchantRequestID string `json:"MerchantRequestID"`
			CheckoutRequestID string `json:"CheckoutRequestID"`
			ResultCode        int    `json:"ResultCode"`
			ResultDesc        string `json:"ResultDesc"`
			CallbackMetadata  struct {
				Item []MetadataItem `json:"Item"`
			} `json:"CallbackMetadata"`
		} `json:"stkCallback"`
	} `json:"Body"`
}

// HandleCallback processes the callback from M-Pesa
func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	var callback CallbackRequest
	if err := json.NewDecoder(r.Body).Decode(&callback); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	log.Println("📥 Callback received")

	body, _ := io.ReadAll(r.Body)
	log.Println("📦 Raw Callback JSON:", string(body))

	data := callback.Body.StkCallback

	// Extract useful fields
	phone := extractMetadataValue(data.CallbackMetadata.Item, "PhoneNumber")
	receipt := extractMetadataValue(data.CallbackMetadata.Item, "MpesaReceiptNumber")
	amount := extractMetadataValue(data.CallbackMetadata.Item, "Amount")
	transactionDate := extractMetadataValue(data.CallbackMetadata.Item, "TransactionDate")

	// Convert if necessary
	dateInt, _ := strconv.ParseInt(transactionDate, 10, 64)

	// Update DB
	_, err := db.DB.Exec(`
		UPDATE stk_requests
		SET status = ?, 
		    mpesa_receipt_number = ?, 
			callback_phone = ?,
		    callback_amount = ?, 
		    transaction_date = ?, 
		    result_code = ?, 
		    result_desc = ?
		WHERE checkout_request_id = ?
	`,
		statusFromCode(data.ResultCode),
		receipt,
		phone,
		amount,
		dateInt,
		data.ResultCode,
		data.ResultDesc,
		data.CheckoutRequestID,
	)

	if err != nil {
		log.Println("❌ Failed to update DB:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Println("✅ Callback handled successfully")
}

// func extractPhoneNumber(items []MetadataItem) string {
// 	for _, item := range items {
// 		if item.Name == "PhoneNumber" {
// 			if phone, ok := item.Value.(string); ok {
// 				return phone
// 			}
// 		}
// 	}
// 	return ""
// }

func extractMetadataValue(items []MetadataItem, key string) string {
	for _, item := range items {
		if item.Name == key {
			switch v := item.Value.(type) {
			case string:
				return v
			case float64:
				return fmt.Sprintf("%.0f", v) // for Amount, TransactionDate
			}
		}
	}
	return ""
}

func statusFromCode(code int) string {
	if code == 0 {
		return "success"
	}
	return "failed"
}

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id") // checkout_request_id
	if id == "" {
		http.Error(w, "Missing id", http.StatusBadRequest)
		return
	}

	var status string
	err := db.DB.QueryRow("SELECT status FROM stk_requests WHERE checkout_request_id = ?", id).Scan(&status)
	if err != nil {
		http.Error(w, "Not found or DB error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": status})
}
