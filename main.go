package main

import (
	"encoding/json"
	"fmt"
	"goRestAPI/db"
	"goRestAPI/mpesa"
	"log"
	"net/http"
	"os"

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

	// res, err := mpesa.InitiateSTKPush(token, req.Phone, req.Amount)
	// if err != nil {
	// 	log.Println("❌ STK Push request error:", err)
	// 	http.Error(w, "Failed to initiate STK push", http.StatusInternalServerError)
	// 	return
	// }

	stkResp, err := mpesa.InitiateSTKPush(token, req.Phone, req.Amount)
	if err != nil {
		log.Println("❌ STK Push request error:", err)
		http.Error(w, "Failed to initiate STK push", http.StatusInternalServerError)
		return
	}
	// defer stkResp.Body.Close()

	// if res.StatusCode != http.StatusOK {
	// 	log.Printf("❌ STK Push failed with status: %s\n", res.Status)
	// 	http.Error(w, "STK Push failed", res.StatusCode)
	// 	return
	// }

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "STK Push initiated successfully",
	})

	_, err = db.DB.Exec(`INSERT INTO stk_requests (phone, amount, status, checkout_request_id) VALUES (?, ?, ?, ?)`,
		req.Phone, req.Amount, "initiated", stkResp.CheckoutRequestID,
	)
	if err != nil {
		log.Println("⚠️ Failed to save to DB:", err)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.Handle("/stkpush", http.HandlerFunc(stkPushHandler))
	http.HandleFunc("/callback", mpesa.CallbackHandler)

	fmt.Printf("🚀 Server running at http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
