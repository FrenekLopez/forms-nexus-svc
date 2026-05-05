package telegram

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/FrenekLopez/forms-nexus/internal/validator"
)

func TestTelegramNotifier_Send_Integration(t *testing.T) {
	// 1. Read environment variables (Meeting Requirement 4)
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")

	// If variables are not set (e.g., when Ilmar runs tests in GitHub),
	// skip the test to avoid breaking the GitHub Actions pipeline.
	if botToken == "" || chatID == "" {
		t.Skip("Skipping integration test: TELEGRAM_BOT_TOKEN or TELEGRAM_CHAT_ID not set")
	}

	// 2. Initialize the provider by injecting credentials
	notifier := &TelegramNotifier{
		BotToken: botToken,
		ChatID:   chatID,
	}

	// 3. Prepare a mock payload for testing
	payload := validator.FormPayload{
		Name:    "Eric",
		Email:   "eric@example.com",
		Message: "Telegram bot built in Go and my Clean Architecture are working",
	}

	// 4. Create a context with a strict 5-second timeout.
	// If Telegram does not respond within 5 seconds, the test aborts.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // Release timer resources after execution

	// 5. Execute the real send operation
	err := notifier.Send(ctx, payload)

	if err != nil {
		t.Fatalf("Failed to send telegram message: %v", err)
	}
}
