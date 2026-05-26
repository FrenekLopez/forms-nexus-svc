package notifier

import (
	"context"
	"errors"

	"github.com/FrenekLopez/forms-nexus/internal/validator"
)

// ErrUnsupportedChannel is returned when an unknown channel is requested.
var ErrUnsupportedChannel = errors.New("unsupported notification channel")

// Notifier defines the strict contract that any provider must fulfill.
type Notifier interface {
	Send(ctx context.Context, payload validator.FormPayload) error
}