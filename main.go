package main

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"math/rand/v2"
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

		slackPayload := buildSlackPayload(data)

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

func buildSlackPayload(data []byte) string {
	icons := []string{
		":fitton-fire:",
		":fitton:",
		":fitton-smile:",
		":fitton-niyari:",
		":fitton-gahaha:",
	}
	randomIcon := icons[rand.IntN(len(icons))]

	var parsed struct {
		AccountID string `json:"account_id"`
	}
	json.Unmarshal(data, &parsed)

	type field struct {
		Title string `json:"title"`
		Value string `json:"value"`
		Short bool   `json:"short"`
	}
	type attachment struct {
		Fields []field `json:"fields"`
	}
	payload := struct {
		Text        string       `json:"text"`
		Attachments []attachment `json:"attachments"`
	}{
		Text: "新しいアカウントができたみたいだよ！" + randomIcon,
		Attachments: []attachment{
			{Fields: []field{
				{Title: "Account ID", Value: parsed.AccountID, Short: true},
			}},
		},
	}
	b, _ := json.Marshal(payload)
	return string(b)
}
