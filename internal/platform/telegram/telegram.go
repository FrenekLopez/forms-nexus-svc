package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"time"

	"github.com/FrenekLopez/forms-nexus/internal/validator"
)

// TelegramNotifier holds the credentials needed to communicate with the Telegram API.
type TelegramNotifier struct {
	BotToken string
	ChatID   string
}

// telegramMessage is a private struct used internally to build the JSON payload.
type telegramMessage struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

// Send formats the form data and sends it as a message to the configured Telegram chat,
// implementing Exponential Backoff and Jitter for network resilience.
func (t *TelegramNotifier) Send(ctx context.Context, payload validator.FormPayload) error {
	// 1. Build a readable message for the user.
	text := fmt.Sprintf("📬 New Form Submission\n\nName: %s\nEmail: %s\nMessage: %s",
		payload.Name, payload.Email, payload.Message)

	// 2. Prepare the message structure for Telegram.
	msg := telegramMessage{
		ChatID: t.ChatID,
		Text:   text,
	}

	// Convert the struct to JSON.
	jsonBody, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal telegram message: %w", err)
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.BotToken)

	// Initialize the HTTP client with a default timeout for serverless environments.
	client := &http.Client{Timeout: 10 * time.Second}

	// --- RETRY CONFIGURATION ---
	maxRetries := 3
	baseDelay := 1 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Create the request inside the loop to ensure a fresh body payload on each attempt.
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBody))
		if err != nil {
			return fmt.Errorf("failed to create http request on attempt %d: %w", attempt, err)
		}
		req.Header.Set("Content-Type", "application/json")

		// Execute the HTTP request.
		resp, err := client.Do(req)

		// 3. Evaluate the success of the request.
		if err == nil && resp.StatusCode == http.StatusOK {
			// Explicitly close the response body and return success.
			resp.Body.Close()
			return nil
		}

		// 4. Handle network or API errors.
		statusCode := 0
		if resp != nil {
			statusCode = resp.StatusCode
			// Explicitly close the body on failure to prevent socket leaks.
			resp.Body.Close()
		}

		// If the maximum number of retries is reached, return the final error.
		if attempt == maxRetries {
			if err != nil {
				return fmt.Errorf("network failure after %d attempts: %w", maxRetries, err)
			}
			return fmt.Errorf("telegram API rejected message after %d attempts, status: %d", maxRetries, statusCode)
		}

		// --- EXPONENTIAL BACKOFF & JITTER ---
		backoffTime := baseDelay * time.Duration(1<<(attempt-1))
		jitter := time.Duration(rand.Intn(500)) * time.Millisecond
		totalWaitTime := backoffTime + jitter

		// Log the temporary failure and retry attempt for observability.
		slog.Warn("Temporary failure sending to Telegram, backing off",
			slog.Int("attempt", attempt),
			slog.Int("max_attempts", maxRetries),
			slog.Int("status_code", statusCode),
			slog.Duration("wait_time", totalWaitTime),
			slog.Any("error", err),
		)

		// 5. Controlled pause: Wait for the calculated backoff time or abort if the context expires.
		select {
		case <-ctx.Done():
			return fmt.Errorf("lambda context expired during backoff wait: %w", ctx.Err())
		case <-time.After(totalWaitTime):
			// Proceed to the next attempt.
		}
	}

	return nil
}
