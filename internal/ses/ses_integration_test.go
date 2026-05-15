package ses_test

import (
	"context"
	"os"
	"testing"

	awsSes "github.com/FrenekLopez/forms-nexus/internal/ses"
	"github.com/FrenekLopez/forms-nexus/internal/validator"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
)

// TestSESIntegration runs a real test against Amazon servers.
// By being named Test... and receiving *testing.T, Go knows it can run it with 'go test'
func TestSESIntegration(t *testing.T) {
	// 1. Read the environment variables configured in PowerShell
	from := os.Getenv("SES_FROM_ADDRESS")
	to := os.Getenv("SES_TO_ADDRESS")

	if from == "" || to == "" {
		t.Skip("Skipping integration test: SES_FROM_ADDRESS or SES_TO_ADDRESS are empty")
	}

	// 2. Load your local AWS configuration (it will read your Windows credentials)
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		t.Fatalf("Critical failure loading AWS Config: %v", err)
	}

	// 3. Prepare the Notifier just like your main.go does
	sesClient := ses.NewFromConfig(cfg)
	notifier := &awsSes.SESNotifier{
		Client:      sesClient,
		FromAddress: from,
		ToAddress:   to,
	}

	// 4. Fabricate a mock form
	payload := validator.FormPayload{
		Name:    "Eric Frenek (Local Test)",
		Email:   "mock-client@test.com",
		Message: "Confirmed.",
	}

	t.Log("Sending request to AWS SES...")
	err = notifier.Send(context.Background(), payload)
	if err != nil {
		t.Fatalf("AWS SES rejected the request: %v", err)
	}

	t.Log("✅ Email sent successfully! Check your inbox.")
}
