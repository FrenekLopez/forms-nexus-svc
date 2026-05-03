package notifier

import (
	"errors"
	"testing"
)

func TestNewNotifierFactory(t *testing.T) {
	// 1. Definimos nuestra "Tabla" de escenarios
	tests := []struct {
		name          string // Nombre de la prueba
		channel       string // El canal que le enviaremos a la fábrica
		expectedError error  // El error que esperamos que nos devuelva
	}{
		{
			name:          "Failure: Unknown channel returns ErrUnsupportedChannel",
			channel:       "sms", // "sms" no existe en nuestro switch
			expectedError: ErrUnsupportedChannel,
		},
		{
			name:          "Failure: Empty channel string returns ErrUnsupportedChannel",
			channel:       "", // Canal vacío tampoco existe
			expectedError: ErrUnsupportedChannel,
		},
		// En el siguiente Sprint, cuando programemos el email, agregaremos esto:
		// {
		// 	name:          "Success: Email channel returns a valid Notifier",
		// 	channel:       "email",
		// 	expectedError: nil,
		// },
	}

	// 2. Iteramos sobre cada escenario de la tabla
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Ejecutamos nuestra fábrica con el canal de prueba
			_, err := NewNotifierFactory(tt.channel)

			// Verificamos si el error recibido es exactamente el que esperábamos
			if !errors.Is(err, tt.expectedError) {
				t.Errorf("expected error %v, got %v", tt.expectedError, err)
			}
		})
	}
}
