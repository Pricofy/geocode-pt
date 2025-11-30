package application

import (
	"encoding/json"
	"testing"

	"github.com/pricofy/geocode-pt/internal/domain"
)

func TestValidateMunicipioOperation(t *testing.T) {
	service := NewPostalCodeService()

	tests := []struct {
		name           string
		event          domain.LambdaEvent
		wantStatusCode int
		wantValid      bool
	}{
		{
			name: "valid municipality",
			event: domain.LambdaEvent{
				Body: `{"operation":"validate-municipio","municipality":"Madrid"}`,
			},
			wantStatusCode: 200,
			wantValid:      true,
		},
		{
			name: "invalid municipality",
			event: domain.LambdaEvent{
				Body: `{"operation":"validate-municipio","municipality":"NonExistentCity"}`,
			},
			wantStatusCode: 200,
			wantValid:      false,
		},
		{
			name: "missing municipality",
			event: domain.LambdaEvent{
				Body: `{"operation":"validate-municipio"}`,
			},
			wantStatusCode: 400,
			wantValid:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := ValidateMunicipalityOperation(service, tt.event)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if response.StatusCode != tt.wantStatusCode {
				t.Errorf("Expected status code %d, got %d", tt.wantStatusCode, response.StatusCode)
			}

			if tt.wantStatusCode == 200 {
				var result domain.ValidationResult
				if err := json.Unmarshal([]byte(response.Body), &result); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if result.Valid != tt.wantValid {
					t.Errorf("Expected valid=%v, got %v", tt.wantValid, result.Valid)
				}
			}
		})
	}
}

func TestAutocompleteMunicipioOperation(t *testing.T) {
	service := NewPostalCodeService()

	tests := []struct {
		name           string
		event          domain.LambdaEvent
		wantStatusCode int
		wantResults    bool
	}{
		{
			name: "valid query",
			event: domain.LambdaEvent{
				Body: `{"operation":"autocomplete-municipio","query":"mad"}`,
			},
			wantStatusCode: 200,
			wantResults:    true,
		},
		{
			name: "valid query with limit",
			event: domain.LambdaEvent{
				Body: `{"operation":"autocomplete-municipio","query":"bar","limit":5}`,
			},
			wantStatusCode: 200,
			wantResults:    true,
		},
		{
			name: "empty query",
			event: domain.LambdaEvent{
				Body: `{"operation":"autocomplete-municipio","query":""}`,
			},
			wantStatusCode: 200,
			wantResults:    true,
		},
		{
			name: "missing query",
			event: domain.LambdaEvent{
				Body: `{"operation":"autocomplete-municipio"}`,
			},
			wantStatusCode: 400,
			wantResults:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := AutocompleteMunicipalityOperation(service, tt.event)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if response.StatusCode != tt.wantStatusCode {
				t.Errorf("Expected status code %d, got %d", tt.wantStatusCode, response.StatusCode)
			}

			if tt.wantStatusCode == 200 {
				var result map[string]interface{}
				if err := json.Unmarshal([]byte(response.Body), &result); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				results, ok := result["results"].([]interface{})
				if !ok {
					t.Error("Expected results array in response")
				}
				if tt.wantResults && len(results) == 0 {
					t.Error("Expected some results, got none")
				}
			}
		})
	}
}

func TestGeocodeByPostalOperation(t *testing.T) {
	service := NewPostalCodeService()

	tests := []struct {
		name           string
		event          domain.LambdaEvent
		wantStatusCode int
		wantSuccess    bool
	}{
		{
			name: "valid postal code",
			event: domain.LambdaEvent{
				Body: `{"operation":"geocode-by-postal","postalCode":"28001"}`,
			},
			wantStatusCode: 200,
			wantSuccess:    true,
		},
		{
			name: "invalid postal code",
			event: domain.LambdaEvent{
				Body: `{"operation":"geocode-by-postal","postalCode":"99999"}`,
			},
			wantStatusCode: 404,
			wantSuccess:    false,
		},
		{
			name: "invalid format",
			event: domain.LambdaEvent{
				Body: `{"operation":"geocode-by-postal","postalCode":"123"}`,
			},
			wantStatusCode: 400,
			wantSuccess:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := GeocodeByPostalOperation(service, tt.event)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if response.StatusCode != tt.wantStatusCode {
				t.Errorf("Expected status code %d, got %d", tt.wantStatusCode, response.StatusCode)
			}
		})
	}
}

func TestReverseGeocodeOperation(t *testing.T) {
	service := NewPostalCodeService()

	tests := []struct {
		name           string
		event          domain.LambdaEvent
		wantStatusCode int
	}{
		{
			name: "valid coordinates",
			event: domain.LambdaEvent{
				Body: `{"operation":"reverse-geocode","lat":40.4168,"lon":-3.7038}`,
			},
			wantStatusCode: 200,
		},
		{
			name: "invalid coordinates",
			event: domain.LambdaEvent{
				Body: `{"operation":"reverse-geocode","lat":999,"lon":999}`,
			},
			wantStatusCode: 400,
		},
		{
			name: "missing coordinates",
			event: domain.LambdaEvent{
				Body: `{"operation":"reverse-geocode"}`,
			},
			wantStatusCode: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := ReverseGeocodeOperation(service, tt.event)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if response.StatusCode != tt.wantStatusCode {
				t.Errorf("Expected status code %d, got %d", tt.wantStatusCode, response.StatusCode)
			}
		})
	}
}

func TestValidatePostalOperation(t *testing.T) {
	service := NewPostalCodeService()

	tests := []struct {
		name           string
		event          domain.LambdaEvent
		wantStatusCode int
		wantValid      bool
	}{
		{
			name: "valid postal code",
			event: domain.LambdaEvent{
				Body: `{"operation":"validate-postal","postalCode":"28001"}`,
			},
			wantStatusCode: 200,
			wantValid:      true,
		},
		{
			name: "invalid postal code",
			event: domain.LambdaEvent{
				Body: `{"operation":"validate-postal","postalCode":"99999"}`,
			},
			wantStatusCode: 200,
			wantValid:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := ValidatePostalOperation(service, tt.event)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if response.StatusCode != tt.wantStatusCode {
				t.Errorf("Expected status code %d, got %d", tt.wantStatusCode, response.StatusCode)
			}

			if tt.wantStatusCode == 200 {
				var result domain.ValidationResult
				if err := json.Unmarshal([]byte(response.Body), &result); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if result.Valid != tt.wantValid {
					t.Errorf("Expected valid=%v, got %v", tt.wantValid, result.Valid)
				}
			}
		})
	}
}

func TestAutocompletePostalOperation(t *testing.T) {
	service := NewPostalCodeService()

	tests := []struct {
		name           string
		event          domain.LambdaEvent
		wantStatusCode int
	}{
		{
			name: "valid prefix",
			event: domain.LambdaEvent{
				Body: `{"operation":"autocomplete-postal","prefix":"28"}`,
			},
			wantStatusCode: 200,
		},
		{
			name: "valid prefix with limit",
			event: domain.LambdaEvent{
				Body: `{"operation":"autocomplete-postal","prefix":"08","limit":5}`,
			},
			wantStatusCode: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := AutocompletePostalOperation(service, tt.event)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if response.StatusCode != tt.wantStatusCode {
				t.Errorf("Expected status code %d, got %d", tt.wantStatusCode, response.StatusCode)
			}
		})
	}
}

// TestErrorHandlers tests the error handling functions
func TestErrorHandlers(t *testing.T) {
	t.Run("handleGeocodeError with PostalCodeNotFoundError", func(t *testing.T) {
		err := domain.NewPostalCodeNotFoundError("28001", "")
		response, _ := handleGeocodeError(err, "Test error")
		if response.StatusCode != 404 {
			t.Errorf("Expected status code 404, got %d", response.StatusCode)
		}
	})

	t.Run("handleGeocodeError with InvalidCoordinatesError", func(t *testing.T) {
		err := domain.NewInvalidCoordinatesError("lat", 999, 999)
		response, _ := handleGeocodeError(err, "Test error")
		if response.StatusCode != 400 {
			t.Errorf("Expected status code 400, got %d", response.StatusCode)
		}
	})

	t.Run("handleGeocodeError with ValidationError", func(t *testing.T) {
		err := domain.NewValidationError("postalCode", "Invalid format")
		response, _ := handleGeocodeError(err, "Test error")
		if response.StatusCode != 400 {
			t.Errorf("Expected status code 400, got %d", response.StatusCode)
		}
	})

	t.Run("handleReverseGeocodeError with InvalidCoordinatesError", func(t *testing.T) {
		err := domain.NewInvalidCoordinatesError("lat", 999, 999)
		response, _ := handleReverseGeocodeError(err, "Test error")
		if response.StatusCode != 400 {
			t.Errorf("Expected status code 400, got %d", response.StatusCode)
		}
	})

	t.Run("handleReverseGeocodeError with PostalCodeNotFoundError", func(t *testing.T) {
		err := domain.NewPostalCodeNotFoundError("", "")
		response, _ := handleReverseGeocodeError(err, "Test error")
		if response.StatusCode != 404 {
			t.Errorf("Expected status code 404, got %d", response.StatusCode)
		}
	})

	t.Run("handleValidationError with ValidationError for autocomplete", func(t *testing.T) {
		err := domain.NewValidationError("prefix", "Invalid prefix")
		response, _ := handleValidationError(err, "autocomplete postal failed")
		if response.StatusCode != 400 {
			t.Errorf("Expected status code 400, got %d", response.StatusCode)
		}
		var result map[string]interface{}
		_ = json.Unmarshal([]byte(response.Body), &result)
		if _, ok := result["results"]; !ok {
			t.Error("Expected results field in response")
		}
	})

	t.Run("handleValidationError with ValidationError for validate", func(t *testing.T) {
		err := domain.NewValidationError("postalCode", "Invalid format")
		response, _ := handleValidationError(err, "validation failed")
		if response.StatusCode != 400 {
			t.Errorf("Expected status code 400, got %d", response.StatusCode)
		}
		var result map[string]interface{}
		_ = json.Unmarshal([]byte(response.Body), &result)
		if _, ok := result["valid"]; !ok {
			t.Error("Expected valid field in response")
		}
	})
}

// TestContainsHelper tests the contains helper function
func TestContainsHelper(t *testing.T) {
	tests := []struct {
		s      string
		substr string
		want   bool
	}{
		{"autocomplete postal", "autocomplete", true},
		{"Autocomplete Postal", "autocomplete", true},
		{"AUTOCOMPLETE", "autocomplete", true},
		{"validate postal", "autocomplete", false},
		{"", "test", false},
		{"test", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.s+"_"+tt.substr, func(t *testing.T) {
			got := contains(tt.s, tt.substr)
			if got != tt.want {
				t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
			}
		})
	}
}

// TestHandleInternalError tests the handleInternalError function
func TestHandleInternalError(t *testing.T) {
	t.Run("with error in dev environment", func(t *testing.T) {
		t.Setenv("ENVIRONMENT", "dev")
		err := domain.NewValidationError("test", "test error")
		response, _ := handleInternalError(err, "Test error")
		if response.StatusCode != 500 {
			t.Errorf("Expected status code 500, got %d", response.StatusCode)
		}
		var result map[string]interface{}
		_ = json.Unmarshal([]byte(response.Body), &result)
		if _, ok := result["details"]; !ok {
			t.Error("Expected details field in dev environment")
		}
	})

	t.Run("with error in production environment", func(t *testing.T) {
		t.Setenv("ENVIRONMENT", "production")
		err := domain.NewValidationError("test", "test error")
		response, _ := handleInternalError(err, "Test error")
		if response.StatusCode != 500 {
			t.Errorf("Expected status code 500, got %d", response.StatusCode)
		}
		var result map[string]interface{}
		_ = json.Unmarshal([]byte(response.Body), &result)
		if _, ok := result["details"]; ok {
			t.Error("Expected no details field in production environment")
		}
	})
}

// TestValidationErrorWithGenericError tests handleValidationError with generic errors
func TestValidationErrorWithGenericError(t *testing.T) {
	t.Run("generic error for autocomplete in dev", func(t *testing.T) {
		t.Setenv("ENVIRONMENT", "dev")
		err := domain.NewValidationError("test", "generic error")
		response, _ := handleValidationError(err, "autocomplete test failed")
		if response.StatusCode != 400 {
			t.Errorf("Expected status code 400, got %d", response.StatusCode)
		}
	})

	t.Run("generic error for validate", func(t *testing.T) {
		err := domain.NewValidationError("test", "generic error")
		response, _ := handleValidationError(err, "validate test failed")
		if response.StatusCode != 400 {
			t.Errorf("Expected status code 400, got %d", response.StatusCode)
		}
	})
}

// TestHandleValidationErrorWithGenericErrorForAutocomplete tests generic error handling for autocomplete
func TestHandleValidationErrorWithGenericErrorForAutocomplete(t *testing.T) {
	t.Run("generic error for autocomplete in dev environment", func(t *testing.T) {
		t.Setenv("ENVIRONMENT", "dev")
		err := domain.NewValidationError("test", "generic error")
		response, _ := handleValidationError(err, "autocomplete test failed")
		if response.StatusCode != 400 {
			t.Errorf("Expected status code 400, got %d", response.StatusCode)
		}
		var result map[string]interface{}
		_ = json.Unmarshal([]byte(response.Body), &result)
		if _, ok := result["results"]; !ok {
			t.Error("Expected results field for autocomplete error")
		}
	})

	t.Run("generic error for autocomplete in production", func(t *testing.T) {
		t.Setenv("ENVIRONMENT", "production")
		err := domain.NewValidationError("test", "generic error")
		response, _ := handleValidationError(err, "autocomplete test failed")
		if response.StatusCode != 400 {
			t.Errorf("Expected status code 400, got %d", response.StatusCode)
		}
	})
}

// TestHandleGeocodeErrorWithGenericError tests generic error handling for geocode
func TestHandleGeocodeErrorWithGenericError(t *testing.T) {
	t.Run("generic error", func(t *testing.T) {
		t.Setenv("ENVIRONMENT", "dev")
		err := domain.NewValidationError("test", "generic error")
		response, _ := handleGeocodeError(err, "Test error")
		if response.StatusCode != 400 {
			t.Errorf("Expected status code 400, got %d", response.StatusCode)
		}
	})
}

// TestHandleReverseGeocodeErrorWithGenericError tests generic error handling for reverse geocode
func TestHandleReverseGeocodeErrorWithGenericError(t *testing.T) {
	t.Run("generic error", func(t *testing.T) {
		t.Setenv("ENVIRONMENT", "dev")
		err := domain.NewValidationError("test", "generic error")
		response, _ := handleReverseGeocodeError(err, "Test error")
		if response.StatusCode != 400 {
			t.Errorf("Expected status code 400, got %d", response.StatusCode)
		}
	})
}

// TestHandleValidationErrorDefaultCaseForAutocomplete tests the default case in handleValidationError for autocomplete
func TestHandleValidationErrorDefaultCaseForAutocomplete(t *testing.T) {
	t.Run("non-validation error for autocomplete in dev", func(t *testing.T) {
		t.Setenv("ENVIRONMENT", "dev")
		// Use a generic error (not ValidationError)
		err := &domain.PostalCodeNotFoundError{}
		response, _ := handleValidationError(err, "autocomplete test failed")
		if response.StatusCode != 500 {
			t.Errorf("Expected status code 500, got %d", response.StatusCode)
		}
		var result map[string]interface{}
		_ = json.Unmarshal([]byte(response.Body), &result)
		if _, ok := result["results"]; !ok {
			t.Error("Expected results field for autocomplete error")
		}
		if _, ok := result["details"]; !ok {
			t.Error("Expected details field in dev environment")
		}
	})

	t.Run("non-validation error for autocomplete in production", func(t *testing.T) {
		t.Setenv("ENVIRONMENT", "production")
		err := &domain.PostalCodeNotFoundError{}
		response, _ := handleValidationError(err, "autocomplete test failed")
		if response.StatusCode != 500 {
			t.Errorf("Expected status code 500, got %d", response.StatusCode)
		}
		var result map[string]interface{}
		_ = json.Unmarshal([]byte(response.Body), &result)
		if _, ok := result["details"]; ok {
			t.Error("Expected no details field in production environment")
		}
	})

	t.Run("non-validation error for validate operation", func(t *testing.T) {
		err := &domain.PostalCodeNotFoundError{}
		response, _ := handleValidationError(err, "validate test failed")
		if response.StatusCode != 500 {
			t.Errorf("Expected status code 500, got %d", response.StatusCode)
		}
		var result map[string]interface{}
		_ = json.Unmarshal([]byte(response.Body), &result)
		if _, ok := result["success"]; !ok {
			t.Error("Expected success field for validate error")
		}
	})
}
