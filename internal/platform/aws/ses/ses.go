package ses

import (
	"context"
	"fmt"

	"github.com/FrenekLopez/forms-nexus/internal/validator"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

type SESAPI interface {
	SendEmail(ctx context.Context, params *ses.SendEmailInput, optFns ...func(*ses.Options)) (*ses.SendEmailOutput, error)
}

type SESNotifier struct {
	Client      SESAPI
	FromAddress string
	ToAddress   string
}

func (s *SESNotifier) Send(ctx context.Context, payload validator.FormPayload) error {
	bodyText := fmt.Sprintf("Nuevo mensaje de contacto:\n\nNombre: %s\nCorreo: %s\nMensaje: %s", payload.Name, payload.Email, payload.Message)

	input := &ses.SendEmailInput{
		Source: aws.String(s.FromAddress),
		Destination: &types.Destination{
			ToAddresses: []string{s.ToAddress},
		},
		Message: &types.Message{
			Subject: &types.Content{
				Data: aws.String("Nueva respuesta de Forms Nexus"),
			},
			Body: &types.Body{
				Text: &types.Content{
					Data: aws.String(bodyText),
				},
			},
		},
	}
	_, err := s.Client.SendEmail(ctx, input)
	if err != nil {
		return fmt.Errorf("ses faile to send email: %w", err)
	}
	return nil
}
