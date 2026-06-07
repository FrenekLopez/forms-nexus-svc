package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"log/slog"

	"github.com/FrenekLopez/forms-nexus/internal/notifier"
	"github.com/FrenekLopez/forms-nexus/internal/platform/aws/dynamodb"
	"github.com/FrenekLopez/forms-nexus/internal/platform/telegram"
	awsSes "github.com/FrenekLopez/forms-nexus/internal/ses"
	"github.com/FrenekLopez/forms-nexus/internal/validator"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/ses"
)

var cwClient *cloudwatch.Client

type App struct {
	EmailNotifier    notifier.Notifier
	TelegramNotifier notifier.Notifier
	DbClient         *dynamodb.Client
}

func init() {
	// Configure the logger to output in JSON format
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Initialize the CloudWatch client here to take advantage of the Cold Start
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		slog.Error("Failed to load AWS initial configuration", slog.String("Error", err.Error()))
	}

	cwClient = cloudwatch.NewFromConfig(cfg)
}

func (a *App) HandlerRequest(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	var payload validator.FormPayload

	slog.Info("Processing new request...")

	if err := json.Unmarshal([]byte(req.Body), &payload); err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400, Body: `{"error":"Invalid JSON"}`}, nil
	}

	if err := payload.Validate(); err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400, Body: `{"error": "Validation failed"}`}, nil
	}

	var activeNotifier notifier.Notifier
	switch payload.TargetChannel {
	case "telegram":
		activeNotifier = a.TelegramNotifier
	case "email":
		activeNotifier = a.EmailNotifier
	default:
		activeNotifier = a.EmailNotifier
	}

	if activeNotifier == nil {
		slog.Error("Selected notifier is not initialized")
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: `{"error": "Notifier not configured"}`}, nil
	}

	if err := activeNotifier.Send(ctx, payload); err != nil {
		slog.Error("Failed to send notification", "Error", err)

		_, cwErr := cwClient.PutMetricData(ctx, &cloudwatch.PutMetricDataInput{
			Namespace: aws.String("FormsNexus"),
			MetricData: []types.MetricDatum{
				{
					MetricName: aws.String("NotificationErrors"),
					Value:      aws.Float64(1.0),
					Unit:       types.StandardUnitCount,
					Dimensions: []types.Dimension{
						{
							Name:  aws.String("TargetChannel"),
							Value: aws.String(payload.TargetChannel),
						},
					},
				},
			},
		})
		if cwErr != nil {
			slog.Error("Failed to publish error metric to CloudWatch", slog.String("Error", cwErr.Error()))
		}

		return events.APIGatewayProxyResponse{StatusCode: 500, Body: `{"error": "Failed to send notification"}`}, nil
	}

	_, cwErr := cwClient.PutMetricData(ctx, &cloudwatch.PutMetricDataInput{
		Namespace: aws.String("FormsNexus"),
		MetricData: []types.MetricDatum{
			{
				MetricName: aws.String("NotificationSuccess"),
				Value:      aws.Float64(1.0),
				Unit:       types.StandardUnitCount,
				Dimensions: []types.Dimension{
					{
						Name:  aws.String("TargetChannel"),
						Value: aws.String(payload.TargetChannel),
					},
				},
			},
		},
	})
	if cwErr != nil {
		slog.Error("Failed to publish successs metric to CloudWatch", slog.String("Error", cwErr.Error()))
	}

	slog.Info("Notification processed successfully", slog.String("Channel", payload.TargetChannel))

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       `{"message": "Form processed successfully!"}`,
	}, nil
}

func main() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	sesClient := ses.NewFromConfig(cfg)
	sesNotifier := &awsSes.SESNotifier{
		Client:      sesClient,
		FromAddress: os.Getenv("SES_FROM_ADDRESS"),
		ToAddress:   os.Getenv("SES_TO_ADDRESS"),
	}

	telegramNotifier := &telegram.TelegramNotifier{
		BotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		ChatID:   os.Getenv("TELEGRAM_CHAT_ID"),
	}

	dbClient, err := dynamodb.NewClient(context.Background())
	if err != nil {
		log.Fatalf("Could not initialize DynamoDB client: &v", err)
	}

	app := &App{
		EmailNotifier:    sesNotifier,
		TelegramNotifier: telegramNotifier,
		DbClient:         dbClient,
	}

	slog.Info("Forms Nexus Service initialized", slog.String("version", "1.1.0"))

	lambda.Start(app.HandlerRequest)
}
