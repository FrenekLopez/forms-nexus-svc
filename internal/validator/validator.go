package validator

import (
	"errors"

	"github.com/go-playground/validator/v10"
)

// We instantiate the validator only once to avoid unnecessary memory consumption on each request.
var validate *validator.Validate

// The init function is executed automatically when the package is imported.
func init() {
	validate = validator.New()
}

// FormPayload represents the incoming data from the web form.
type FormPayload struct {
	Name          string `json:"name" validate:"required"`
	Email         string `json:"email" validate:"required,email"`
	Message       string `json:"message" validate:"required"`
	TargetChannel string `json:"target_channel"`
}

// Validate evaluates the struct and returns clear,custom Go errors as required by the ticket.
func (f *FormPayload) Validate() error {
	err := validate.Struct(f)

	if err != nil {
		// We cast the error to validator.ValidationErrors to access specific field metadate.
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			firstErr := validationErrors[0]

			switch firstErr.Field() {
			case "Name":
				return errors.New("the field 'name' is required")
			case "Email":
				if firstErr.Tag() == "email" {
					return errors.New("the field 'email' is invalid")
				}
				return errors.New("the field 'email' required")
			case "Message":
				return errors.New("the field 'message' is required")

			}
		}
		return err
	}
	return nil
}
