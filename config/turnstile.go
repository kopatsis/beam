package config

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
)

type TurnstileResponse struct {
	Success bool `json:"success"`
}

func VerifyTurnstile(client *http.Client, token string) (bool, error) {
	secretKey := os.Getenv("TURNSTILE_SECRET")
	url := "https://challenges.cloudflare.com/turnstile/v0/siteverify"

	data := map[string]string{
		"secret":   secretKey,
		"response": token,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return false, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return false, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var result TurnstileResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	return result.Success, nil
}
