package mpesa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
    "encoding/base64"
	"github.com/skip2/go-qrcode"
)

type QRCode struct {
	Content string
	Size    int
}

type QRCodeResponse struct {
	ResponseCode        string `json:"ResponseCode"`
	CheckoutRequestID   string `json:"CheckoutRequestID"`
	ResponseDescription string `json:"ResponseDescription"`
	QRCode              string `json:"QRCode"` // Base64 encoded QR code image
}

type QRCodeRequest struct {
	CPI     string `json:"cpi"`
	Amount  string `json:"amount"`         // Received as string, parsed inside handler
	TrxCode string `json:"transaction_code"`
}

func (code *QRCode) Generate() ([]byte, error) {
	qrCode, err := qrcode.Encode(code.Content, qrcode.Medium, code.Size)
	if err != nil {
		return nil, fmt.Errorf("could not generate QR code: %v", err)
	}
	return qrCode, nil
}

func QRCodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req QRCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.CPI == "" || req.Amount == "" || req.TrxCode == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	amountFloat, err := strconv.ParseFloat(req.Amount, 64)
	if err != nil {
		http.Error(w, "Invalid amount format", http.StatusBadRequest)
		return
	}

	tm := &TokenManager{}
	token, err := tm.GetToken()
	if err != nil {
		log.Println("❌ Failed to get token:", err)
		http.Error(w, "Failed to get token", http.StatusInternalServerError)
		return
	}

	qrResp, err := InitiateQRCode(req.CPI, amountFloat, req.TrxCode, token)
	if err != nil {
		log.Println("❌ QR Code generation failed:", err)
		http.Error(w, "Failed to initiate dynamic QR code", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(qrResp)
}

func InitiateQRCode(cpi string, amount float64, trxCode string, token string) (*QRCodeResponse, error) {
	payload := map[string]interface{}{
		"MerchantName": "TEST ORUCHO",
		"RefNo":        "Invoice Test",
		"CPI":          cpi,
		"Amount":       amount,
		"TrxCode":      trxCode,
		"Size":         "300",
	}

	body, err := json.Marshal(payload)
	log.Println(payload)
	log.Println(string(body))	
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %v", err)
	}

	req, err := http.NewRequest("POST", "https://sandbox.safaricom.co.ke/mpesa/qrcode/v1/generate", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	log.Println("📦 Daraja QR Response:", string(respBody))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("daraja api error: %s", string(respBody))
	}

	var qrResp QRCodeResponse
	if err := json.Unmarshal(respBody, &qrResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &qrResp, nil
}

func QRCodeImageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req QRCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.CPI == "" || req.Amount == "" || req.TrxCode == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	amountFloat, err := strconv.ParseFloat(req.Amount, 64)
	if err != nil {
		http.Error(w, "Invalid amount format", http.StatusBadRequest)
		return
	}

	tm := &TokenManager{}
	token, err := tm.GetToken()
	if err != nil {
		log.Println("❌ Failed to get token:", err)
		http.Error(w, "Failed to get token", http.StatusInternalServerError)
		return
	}

	qrResp, err := InitiateQRCode(req.CPI, amountFloat, req.TrxCode, token)
	if err != nil {
		log.Println("❌ QR Code generation failed:", err)
		http.Error(w, "Failed to initiate dynamic QR code", http.StatusInternalServerError)
		return
	}

	// Decode Base64 string to image bytes
	imgBytes, err := decodeBase64Image(qrResp.QRCode)
	if err != nil {
		log.Println("❌ Failed to decode QR code image:", err)
		http.Error(w, "Failed to decode QR code image", http.StatusInternalServerError)
		return
	}

	// Set content type and write image
	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(http.StatusOK)
	w.Write(imgBytes)
}

func decodeBase64Image(data string) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 image: %v", err)
	}
	return decoded, nil
}

