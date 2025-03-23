package logging

import (
	"beam/background/emails"
	"beam/config"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func LogsToLoggly(tools *config.Tools, payload []byte) {
	go func() {
		token := os.Getenv("LOGGLY_TOKEN")
		req, err := http.NewRequest("POST", "https://logs-01.loggly.com/inputs/"+token+"/tag/http/", bytes.NewReader(payload))
		if err != nil {
			emails.BackupLogEmail("Unable to push non-critical log(s) to Loggly", string(payload), err, tools)
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := tools.Client.Do(req)
		if err != nil {
			emails.BackupLogEmail("Unable to push non-critical log(s) to Loggly", string(payload), err, tools)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			emails.BackupLogEmail("Unable to push non-critical log(s) to Loggly", string(payload), fmt.Errorf("failed to send logs, status code: %d", resp.StatusCode), tools)
		}
	}()
}

func AsyncCriticalError(tools *config.Tools, logID, description string) {
	go func() {
		log := gin.H{
			"level":         "Error",
			"associated_id": logID,
			"description":   description,
		}

		payload, err := json.Marshal(log)
		if err != nil {
			payload = []byte(`{"level":"Error","associated_id":"` + logID + `","description":"` + description + `"}`)
		}

		token := os.Getenv("LOGGLY_TOKEN")
		req, err := http.NewRequest("POST", "https://logs-01.loggly.com/inputs/"+token+"/tag/http/", bytes.NewReader(payload))
		if err != nil {
			emails.BackupLogEmail("Unable to push critical error to Loggly", string(payload), err, tools)
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := tools.Client.Do(req)
		if err != nil {
			emails.BackupLogEmail("Unable to push critical error to Loggly", string(payload), err, tools)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			emails.BackupLogEmail("Unable to push critical error to Loggly", string(payload), fmt.Errorf("failed to send logs, status code: %d", resp.StatusCode), tools)
		}
	}()
}

func HeartBeat(tools *config.Tools) {
	payload := []byte(`{"level":"Informational","Success":true}`)

	token := os.Getenv("LOGGLY_TOKEN")
	req, err := http.NewRequest("POST", "https://logs-01.loggly.com/inputs/"+token+"/tag/http/", bytes.NewReader(payload))
	if err != nil {
		emails.BackupLogEmail("Unable to push heartbeat message to Loggly", string(payload), err, tools)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := tools.Client.Do(req)
	if err != nil {
		emails.BackupLogEmail("Unable to push heartbeat message to Loggly", string(payload), err, tools)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		emails.BackupLogEmail("Unable to push heartbeat message to Loggly", string(payload), fmt.Errorf("failed to send logs, status code: %d", resp.StatusCode), tools)
	}
}
