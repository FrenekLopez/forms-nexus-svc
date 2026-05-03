package notifier

import (
	"context"
	"errors"

	"github.com/FrenekLopez/forms-nexus/internal/validator"
)

// ErrUnsupportedChannel is returned when the factory receives an unknown channel.
var ErrUnsupportedChannel = errors.New("unsupported notification channel")

// Notifier defines the strict contract that any provider must fulfill.
type Notifier interface {
	Send(ctx context.Context, payload validator.FormPayload) error
}

// NewNotifierFactory acts as a router that decides which sender engine to instantiate.
func NewNotifierFactory(channel string) (Notifier, error) {
	switch channel {
	// Future implementations:
	// case "email":
	//     return &AWSSESNotifier{}, nil
	// case "telegram":
	//     return &TelegramNotifier{}, nil
	default:
		// Protect the system from unknown channels by returning a controlled error.
		return nil, ErrUnsupportedChannel
	}
}
