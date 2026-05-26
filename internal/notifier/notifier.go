package notifier

import (
	"context"

	"github.com/FrenekLopez/forms-nexus/internal/validator"
)

// Notifier defines the strict contract that any provider must fulfill.
type Notifier interface {
	Send(ctx context.Context, payload validator.FormPayload) error
}
