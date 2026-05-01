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
	// Configuramos el logger para que la salida sea en formato JSON
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	slog.Info("Procesando nueva peticion",
		slog.String("http_method", req.HTTPMethod),
		slog.String("path", req.Path),
	)
	var payload validator.FormPayload

	// Combertimos el JSON (que bieene como texto en req.Body) a nuestro struct.
	err := json.Unmarshal([]byte(req.Body), &payload)
	if err != nil {
		slog.Error("FAllo al procesar el JSON (posible ataque o malformato)",
			slog.String("error", err.Error()),
		)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadGateway,
			Body:       `{"error": "El formato del JSON es invalido"}`,
		}, nil
	}

	// Llamamos la logica del Spring 1
	err = payload.Validate()
	if err != nil {
		slog.Error("Validacion de campos fallida",
			slog.String("error", err.Error()),
			slog.String("email_intentado", payload.Email),
		)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       `{"error": "` + err.Error() + `"}`,
		}, nil
	}
	slog.Info("Formulario validado exitosamente", slog.String("email", payload.Email))

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       `{"message": "Formulario procesado correctamente"}`,
	}, nil
}

func main() {
	lambda.Start(Handler)
}
