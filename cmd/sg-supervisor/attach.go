package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

func existingSupervisorURL(listen string) (string, bool) {
	host, port, err := net.SplitHostPort(listen)
	if err != nil {
		return "", false
	}
	if host == "" || host == "0.0.0.0" || host == "::" {
		host = "127.0.0.1"
	}
	baseURL := fmt.Sprintf("http://%s:%s", host, port)
	client := &http.Client{Timeout: 700 * time.Millisecond}

	response, err := client.Get(baseURL + "/api/v1/status")
	if err != nil {
		return "", false
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", false
	}
	var body struct {
		Success bool `json:"success"`
		Data    struct {
			ProductName string `json:"productName"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		return "", false
	}
	if !body.Success || !strings.EqualFold(body.Data.ProductName, "School Gate") {
		return "", false
	}
	return baseURL, true
}
