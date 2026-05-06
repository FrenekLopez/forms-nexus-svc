package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	// Make sure this path matches your module name in go.mod
	"github.com/FrenekLopez/forms-nexus/internal/validator"
)

// TelegramNotifier holds the credentials needed to communicate with the Telegram API.
type TelegramNotifier struct {
	BotToken string
	ChatID   string
}

// telegramMessage is a private struct used internally to build the JSON payload
// required by the Telegram Bot API.
type telegramMessage struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

// Send formats the form data and sends it as a message to the configured Telegram chat.
func (t *TelegramNotifier) Send(ctx context.Context, payload validator.FormPayload) error {
	// 1. Build a readable message for the user
	text := fmt.Sprintf("📬 New Form Submission\n\nName: %s\nEmail: %s\nMessage: %s",
		payload.Name, payload.Email, payload.Message)

	// 2. Prepare the message structure for Telegram
	msg := telegramMessage{
		ChatID: t.ChatID,
		Text:   text,
	}

	// Convert the struct to JSON
	jsonBody, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal telegram message: %w", err)
	}

	// 3. Configure the HTTP request
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.BotToken)

	// Create the request with context for proper cancellation and timeout support
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create http request: %w", err)
	}

	// Set the appropriate content type
	req.Header.Set("Content-Type", "application/json")

	// 4. Execute the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to telegram: %w", err)
	}

	// Always close the response body to prevent resource leaks
	defer resp.Body.Close()

	// 5. Check if the request was successful
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned status: %d", resp.StatusCode)
	}

	return nil
}
