package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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

func (a *App) HandlerRequest(ctx context.Context, sqsEvent events.SQSEvent) error {

	for _, record := range sqsEvent.Records {
		var payload validator.FormPayload

		slog.Info("Processing new message from SQS queue...", slog.String("MessageId", record.MessageId))

		if err := json.Unmarshal([]byte(record.Body), &payload); err != nil {
			slog.Error("Invalid JSON in SQS message", slog.String("Error", err.Error()))
			// Si el JSON viene mal, retornamos el error. SQS lo intentará 3 veces y luego lo mandará a la DLQ.
			return fmt.Errorf("invalid json: %w", err)
		}

		if err := payload.Validate(); err != nil {
			slog.Error("Validation failed for SQS message", slog.String("Error", err.Error()))
			return fmt.Errorf("validation failed: %w", err)
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
			return fmt.Errorf("Notifier not configured")
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

			return fmt.Errorf("Failed to send notification: %w", err)
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

	}
	return nil
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
		log.Fatalf("Could not initialize DynamoDB client: %v", err)
	}

	app := &App{
		EmailNotifier:    sesNotifier,
		TelegramNotifier: telegramNotifier,
		DbClient:         dbClient,
	}

	slog.Info("Forms Nexus Service initialized", slog.String("version", "1.1.0"))

	lambda.Start(app.HandlerRequest)
}
