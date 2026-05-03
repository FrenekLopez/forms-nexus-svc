package validator

import (
	"testing"
)

func TestFormPayload_Validate(t *testing.T) {
	// Define our "table" of test cases
	tests := []struct {
		name    string      // Descriptive test name
		payload FormPayload // Mock data to inject
		wantErr bool        // Do we expect an error? (true/false)
	}{
		{
			name: "Success: Fully valid payload",
			payload: FormPayload{
				Name:          "Eric",
				Email:         "eric@example.com",
				Message:       "I want information about the service.",
				TargetChannel: "email",
			},
			wantErr: false, // We do NOT expect an error
		},
		{
			name: "Failure: Empty email",
			payload: FormPayload{
				Name:    "Eric",
				Email:   "", // Empty field
				Message: "Hello",
			},
			wantErr: true, // We DO expect an error
		},
		{
			name: "Failure: Invalid email format",
			payload: FormPayload{
				Name:    "Eric",
				Email:   "fake-email-without-at", // Incorrect format
				Message: "Hello",
			},
			wantErr: true, // We DO expect an error
		},
		{
			name: "Failure: Missing message",
			payload: FormPayload{
				Name:    "Eric",
				Email:   "eric@example.com",
				Message: "", // Empty field
			},
			wantErr: true, // We DO expect an error
		},
	}

	// Execute each test case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.payload.Validate()

			// Check if the result matches expectations
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() returned error = %v, but expected wantErr = %v", err, tt.wantErr)
			}
		})
	}
}
