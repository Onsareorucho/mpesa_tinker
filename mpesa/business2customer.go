package mpesa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// {
//    "OriginatorConversationID": "feb5e3f2-fbbc-4745-844c-ee37b546f627",
//    "InitiatorName": "testapi",
//    "SecurityCredential":"EsJocK7+NjqZPC3I3EO+TbvS+xVb9TymWwaKABoaZr/Z/n0UysSs..",
//    "CommandID":"BusinessPayment",
//    "Amount":"10"
//    "PartyA":"600996",
//    "PartyB":"254728762287"
//    "Remarks":"here are my remarks",
//    "QueueTimeOutURL":"https://mydomain.com/b2c/queue",
//    "ResultURL":"https://mydomain.com/b2c/result",
//    "Occassion":"Christmas"
// }

type BusinessToCustomerRequest struct {
	OriginatorConversationID string `json:"OriginatorConversationID"`
	InitiatorName            string `json:"InitiatorName"`
	SecurityCredential       string `json:"SecurityCredential"`
	CommandID                string `json:"CommandID"`
	Amount                   int    `json:"Amount"`
	PartyA                   int    `json:"PartyA"` // Shortcode or phone number
	PartyB                   int    `json:"PartyB"` // Phone number of the recipient
	Remarks                  string `json:"Remarks"`
	QueueTimeOutURL          string `json:"QueueTimeOutURL"` // URL to call if the request times out
	ResultURL                string `json:"ResultURL"`       // URL to call with the result of the transaction
	Occassion                string `json:"Occassion"`       // Optional, can be used for additional information
}

type BusinessToCustomerResponse struct {
	ConversationID           string `json:"ConversationID"`
	OriginatorConversationID string `json:"OriginatorConversationID"`
	ResponseCode             string `json:"ResponseCode"`
	ResponseDescription      string `json:"ResponseDescription"`
}

// {
//  "ConversationID": "AG_20191219_00005797af5d7d75f652",
//  "OriginatorConversationID": "16740-34861180-1",
//  "ResponseCode": "0",
//  "ResponseDescription": "Accept the service request successfully."
// }

func BusinessToCustomerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req BusinessToCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.OriginatorConversationID == "" ||
		req.InitiatorName == "" || req.SecurityCredential == "" || req.CommandID == "" {
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

	B2CResp, err := InitiateB2C(token, req)
	if err != nil {
		log.Println("❌ B2C error:", err)
		http.Error(w, "Failed to make the B2C transaction", http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(B2CResp)

}

func InitiateB2C(token string, req BusinessToCustomerRequest) (*BusinessToCustomerResponse, error) {
	// 1. marshal the request payload
	payload := map[string]interface{}{
		"OriginatorConversationID": req.OriginatorConversationID,
		"InitiatorName":            req.InitiatorName,
		"SecurityCredential":       req.SecurityCredential,
		"CommandID":                req.CommandID,
		"Amount":                   req.Amount,
		"PartyA":                   req.PartyA,
		"PartyB":                   req.PartyB,
		"Remarks":                  req.Remarks,
		"QueueTimeOutURL":          req.QueueTimeOutURL,
		"ResultURL":                req.ResultURL,
		"Occassion":                req.Occassion,
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	// 2. build the HTTP request
	httpReq, err := http.NewRequest(
		http.MethodPost,
		"https://sandbox.safaricom.co.ke/mpesa/b2c/v3/paymentrequest",
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")

	// 3. perform the call
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	// 4. read the body once
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	log.Println("📦 Daraja B2C response:", string(respBytes))

	// 5. optional: check non-200 status codes early
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("daraja returned %d: %s", resp.StatusCode, string(respBytes))
	}

	// 6. unmarshal from the bytes we already read
	var b2cResp BusinessToCustomerResponse
	if err := json.Unmarshal(respBytes, &b2cResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &b2cResp, nil
}

