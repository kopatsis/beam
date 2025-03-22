package logging

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"os"
)

func LogsToLoggly(payload []byte, client *http.Client) error {
	token := os.Getenv("LOGGGLY_TOKEN")
	if token == "" {
		return errors.New("no loggly token")
	}

	req, err := http.NewRequest("POST", "https://logs-01.loggly.com/inputs/"+token+"/tag/http/", bytes.NewReader(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send logs, status code: %d", resp.StatusCode)
	}

	return nil
}
