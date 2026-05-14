package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"log/slog"

	"github.com/FrenekLopez/forms-nexus/internal/notifier"
	"github.com/FrenekLopez/forms-nexus/internal/validator"

	awsSes "github.com/FrenekLopez/forms-nexus/internal/platform/aws/ses"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
)

type App struct {
	Notifier notifier.Notifier
}

func init() {
	// Configure the logger to output in JSON format
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
}

func (a *App) HandlerRequest(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	var payload validator.FormPayload

	if err := json.Unmarshal([]byte(req.Body), &payload); err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400, Body: `{"error":"Invalid JSON"}`}, nil
	}

	if err := payload.Validate(); err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400, Body: `{"error": "Validation failed"}`}, nil
	}

	if err := a.Notifier.Send(ctx, payload); err != nil {
		slog.Error("Failed to send notification", "Error", err)
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: `{"error": "Failed to send email"}`}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       `{"message": "Form processed and email sent!"}`,
	}, nil
}

func main() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("No se puede cargar AWS config: %v", err)
	}

	sesClient := ses.NewFromConfig(cfg)

	sesNotifier := &awsSes.SESNotifier{
		Client:      sesClient,
		FromAddress: os.Getenv("SES_FROM_ADDRESS"),
		ToAddress:   os.Getenv("SES_TO_ADDRESS"),
	}

	app := &App{
		Notifier: sesNotifier,
	}

	lambda.Start(app.HandlerRequest)
}
