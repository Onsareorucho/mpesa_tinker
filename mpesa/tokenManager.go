package mpesa

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"sync"
	"time"
)

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   string `json:"expires_in"`
}

type TokenManager struct {
	Token     string
	ExpiresAt time.Time
	Mu        sync.Mutex
}

func (tm *TokenManager) GetToken() (string, error) {
	tm.Mu.Lock()
	defer tm.Mu.Unlock()

	if time.Now().Before(tm.ExpiresAt) && tm.Token != "" {
		return tm.Token, nil
	}

	consumerKey := os.Getenv("MPESA_CONSUMER_KEY")
	consumerSecret := os.Getenv("MPESA_CONSUMER_SECRET")

	url := "https://sandbox.safaricom.co.ke/oauth/v1/generate?grant_type=client_credentials"
	req, _ := http.NewRequest("GET", url, nil)

	credentials := base64.StdEncoding.EncodeToString([]byte(consumerKey + ":" + consumerSecret))
	req.Header.Add("Authorization", "Basic "+credentials)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", errors.New("failed to get token")
	}

	var tokenResp TokenResponse
	err = json.NewDecoder(res.Body).Decode(&tokenResp)
	if err != nil {
		return "", err
	}

	duration, _ := time.ParseDuration(tokenResp.ExpiresIn + "s")
	tm.Token = tokenResp.AccessToken
	tm.ExpiresAt = time.Now().Add(duration)

	return tm.Token, nil
}
