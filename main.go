package main

import (
	"encoding/json"
	"fmt"
	"goRestAPI/mpesa"
	"log"
	"net/http"
	"os"
	"goRestAPI/db"
	"github.com/joho/godotenv"
)

type STKRequest struct {
	Phone  string `json:"phone"`
	Amount string `json:"amount"`
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  Warning: No .env file found. Environment variables must be set manually.")
	}
	db.CreateDatabaseAndTables()
}

func stkPushHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req STKRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Phone == "" || req.Amount == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tm := &mpesa.TokenManager{}
	token, err := tm.GetToken()
	if err != nil {
		log.Println("❌ Failed to get token:", err)
		http.Error(w, "Failed to get token", http.StatusInternalServerError)
		return
	}

	res, err := mpesa.InitiateSTKPush(token, req.Phone, req.Amount)
	if err != nil {
		log.Println("❌ STK Push request error:", err)
		http.Error(w, "Failed to initiate STK push", http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Printf("❌ STK Push failed with status: %s\n", res.Status)
		http.Error(w, "STK Push failed", res.StatusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "STK Push initiated successfully",
	})

	_, err = db.DB.Exec(`INSERT INTO stk_requests (phone, amount, status) VALUES (?, ?, ?)`,
		req.Phone, req.Amount, "initiated",
	)
	if err != nil {
		log.Println("⚠️ Failed to save to DB:", err)
	}
}

// func callbackHandler(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodPost {
// 		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	var callback struct {
// 		Body struct {
// 			StkCallback struct {
// 				MerchantRequestID string `json:"MerchantRequestID"`
// 				CheckoutRequestID string `json:"CheckoutRequestID"`
// 				ResultCode        int    `json:"ResultCode"`
// 				ResultDesc        string `json:"ResultDesc"`
// 			} `json:"stkCallback"`
// 		} `json:"Body"`
// 	}

// 	err := json.NewDecoder(r.Body).Decode(&callback)
// 	if err != nil {
// 		log.Println("❌ Failed to decode callback:", err)
// 		http.Error(w, "Invalid request", http.StatusBadRequest)
// 		return
// 	}

// 	stk := callback.Body.StkCallback

// 	status := "unknown"
// 	if stk.ResultCode == 0 {
// 		status = "success"
// 	} else {
// 		status = "failed"
// 	}

// 	_, err = DB.Exec(`UPDATE stk_requests SET status = ? WHERE checkout_request_id = ?`, status, stk.CheckoutRequestID)
// 	if err != nil {
// 		log.Println("❌ Failed to update transaction:", err)
// 		http.Error(w, "DB update error", http.StatusInternalServerError)
// 		return
// 	}

// 	log.Printf("✅ Callback received for %s: %s\n", stk.CheckoutRequestID, status)
// 	w.WriteHeader(http.StatusOK)
// }

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.Handle("/stkpush", http.HandlerFunc(stkPushHandler))

	fmt.Printf("🚀 Server running at http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
