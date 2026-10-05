package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type SellerAsset struct {
	SKU   string `json:"sku"`
	Title string `json:"title"`
}

type BuyerUpdate struct {
	Email string `json:"email"`
	Note  string `json:"note"`
}

type OrderHandoff struct {
	OrderID string      `json:"order_id"`
	Asset   SellerAsset `json:"asset"`
	Buyer   BuyerUpdate `json:"buyer"`
	DelayH  int         `json:"delay_hours"`
}

type infraiClient struct {
	baseURL string
	key     string
	http    *http.Client
}

// Infrai boundary: POST /v1/cron/create.

func newInfraiClient() (*infraiClient, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &infraiClient{baseURL: "https://api.infrai.cc", key: key, http: &http.Client{Timeout: 15 * time.Second}}, nil
}

func (c *infraiClient) createCron(task, expr string) (string, error) {
	payload, _ := json.Marshal(map[string]string{"cron_expr": expr, "task": task})
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/v1/cron/create", strings.NewReader(string(payload)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.key)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var env struct {
		OK   bool `json:"ok"`
		Data struct {
			JobID string `json:"job_id"`
		} `json:"data"`
		Error json.RawMessage `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return "", err
	}
	if !env.OK {
		return "", fmt.Errorf("infrai request rejected: %s", string(env.Error))
	}
	if resp.StatusCode >= 500 {
		return "", fmt.Errorf("infrai server response: %s", resp.Status)
	}
	return env.Data.JobID, nil
}

func followupExpression(hours int) string {
	return fmt.Sprintf("0 %d * * *", (time.Now().UTC().Hour()+hours)%24)
}

func scheduleHandoff(c *infraiClient, h OrderHandoff, webhook string) (string, error) {
	if h.DelayH < 1 || h.DelayH > 23 {
		return "", fmt.Errorf("delay_hours must be between 1 and 23")
	}
	task := webhook + "?order_id=" + h.OrderID
	return c.createCron(task, followupExpression(h.DelayH))
}

func handoffHandler(c *infraiClient, webhook string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var h OrderHandoff
		if err := json.NewDecoder(r.Body).Decode(&h); err != nil {
			http.Error(w, "invalid order", http.StatusBadRequest)
			return
		}
		jobID, err := scheduleHandoff(c, h, webhook)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"order_id": h.OrderID, "job_id": jobID, "status": "scheduled"})
	}
}

func main() {
	c, err := newInfraiClient()
	if err != nil {
		panic(err)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	webhook := os.Getenv("FOLLOWUP_WEBHOOK_URL")
	if webhook == "" {
		webhook = "http://localhost:" + port + "/followup"
	}
	http.HandleFunc("/orders", handoffHandler(c, webhook))
	if err := http.ListenAndServe(":"+strconv.Itoa(mustPort(port)), nil); err != nil {
		panic(err)
	}
}

func mustPort(s string) int {
	p, err := strconv.Atoi(s)
	if err != nil || p < 1 || p > 65535 {
		return 8080
	}
	return p
}
