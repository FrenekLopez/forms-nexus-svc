package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Obtenerl la URL de la cola que inyectamos desde CDK
	queueURL := os.Getenv("QUEUE_URL")

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Printf("Error logging AWS config: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, nil
	}

	sqsClient := sqs.NewFromConfig(cfg)

	output, err := sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(queueURL),
		MessageBody: aws.String(req.Body),
	})
	if err != nil {
		log.Printf("Error sending message to SQS: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, nil
	}

	log.Printf("Successfully queued message with ID: %s", *output.MessageId)

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       `{"message": "Form successfully queue for processing"}`,
		Headers: map[string]string{
			"Content-Type": "application json",
		},
	}, nil
}

func main() {
	lambda.Start(handler)
}
