package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"log/slog"

	"github.com/FrenekLopez/forms-nexus/internal/validator"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func init() {
	// Configure the logger to output in JSON format
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	slog.Info("Processing new request",
		slog.String("http_method", req.HTTPMethod),
		slog.String("path", req.Path),
	)

	var payload validator.FormPayload

	// Convert the JSON (which comes as a string in req.Body) into our struct
	err := json.Unmarshal([]byte(req.Body), &payload)
	if err != nil {
		slog.Error("Failed to process JSON (possible malformed input or attack)",
			slog.String("error", err.Error()),
		)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       `{"error": "Invalid JSON format"}`,
		}, nil
	}

	// Call validation logic (Sprint 1)
	err = payload.Validate()
	if err != nil {
		slog.Error("Field validation failed",
			slog.String("error", err.Error()),
			slog.String("attempted_email", payload.Email),
		)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       `{"error": "` + err.Error() + `"}`,
		}, nil
	}

	slog.Info("Form successfully validated", slog.String("email", payload.Email))

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       `{"message": "Form processed successfully"}`,
	}, nil
}

func main() {
	lambda.Start(Handler)
}
