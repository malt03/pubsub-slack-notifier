package main

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type pubSubMessage struct {
	Message struct {
		Data string `json:"data"`
	} `json:"message"`
}

func main() {
	slackWebhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	if slackWebhookURL == "" {
		log.Fatal("SLACK_WEBHOOK_URL environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("failed to read body: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		var msg pubSubMessage
		if err := json.Unmarshal(body, &msg); err != nil {
			log.Printf("failed to unmarshal pubsub message: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		data, err := base64.StdEncoding.DecodeString(msg.Message.Data)
		if err != nil {
			log.Printf("failed to decode base64 data: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		text := formatMessage(data)
		slackPayload := buildSlackPayload(text)

		resp, err := http.Post(slackWebhookURL, "application/json", strings.NewReader(slackPayload))
		if err != nil {
			log.Printf("failed to post to slack: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(resp.Body)
			log.Printf("slack returned %d: %s", resp.StatusCode, string(respBody))
			http.Error(w, "slack error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	log.Printf("starting server on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func buildSlackPayload(text string) string {
	payload := map[string]string{"text": text}
	b, _ := json.Marshal(payload)
	return string(b)
}

func formatMessage(data []byte) string {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return string(data)
	}
	pretty, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return string(data)
	}
	return "```\n" + string(pretty) + "\n```"
}
