//nolint:dupl // test code duplication is acceptable for clarity
package handler

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/pricofy/geocode-pt/internal/domain"
)

func TestHandler_GeocodeByPostal(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		event          domain.LambdaEvent
		wantStatus     int
		wantSuccess    bool
		wantPostalCode string
	}{
		{
			name: "valid postal code",
			event: domain.LambdaEvent{
				Body: `{"operation":"geocode-by-postal","postalCode":"28001"}`,
			},
			wantStatus:     200,
			wantSuccess:    true,
			wantPostalCode: "28001",
		},
		{
			name: "valid municipality",
			event: domain.LambdaEvent{
				Body: `{"operation":"geocode-by-postal","municipality":"Madrid"}`,
			},
			wantStatus:  200,
			wantSuccess: true,
		},
		{
			name: "invalid postal code",
			event: domain.LambdaEvent{
				Body: `{"operation":"geocode-by-postal","postalCode":"99999"}`,
			},
			wantStatus:  404,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := Handler(ctx, tt.event)

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if response.StatusCode != tt.wantStatus {
				t.Errorf("Expected status code %d, got %d", tt.wantStatus, response.StatusCode)
			}

			var body map[string]interface{}
			if err := json.Unmarshal([]byte(response.Body), &body); err != nil {
				t.Fatalf("Failed to parse response body: %v", err)
			}

			if success, ok := body["success"].(bool); ok && success != tt.wantSuccess {
				t.Errorf("Expected success=%v, got %v", tt.wantSuccess, success)
			}

			if tt.wantPostalCode != "" {
				if postalCode, ok := body["postalCode"].(string); ok && postalCode != tt.wantPostalCode {
					t.Errorf("Expected postalCode=%s, got %s", tt.wantPostalCode, postalCode)
				}
			}
		})
	}
}

func TestHandler_ReverseGeocode(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		event       domain.LambdaEvent
		wantStatus  int
		wantSuccess bool
	}{
		{
			name: "valid coordinates",
			event: domain.LambdaEvent{
				Body: `{"operation":"reverse-geocode","lat":40.4168,"lon":-3.7038}`,
			},
			wantStatus:  200,
			wantSuccess: true,
		},
		{
			name: "missing coordinates",
			event: domain.LambdaEvent{
				Body: `{"operation":"reverse-geocode"}`,
			},
			wantStatus:  400,
			wantSuccess: false,
		},
		{
			name: "invalid coordinates",
			event: domain.LambdaEvent{
				Body: `{"operation":"reverse-geocode","lat":100,"lon":-3.7038}`,
			},
			wantStatus:  400,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := Handler(ctx, tt.event)

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if response.StatusCode != tt.wantStatus {
				t.Errorf("Expected status code %d, got %d", tt.wantStatus, response.StatusCode)
			}

			var body map[string]interface{}
			if err := json.Unmarshal([]byte(response.Body), &body); err != nil {
				t.Fatalf("Failed to parse response body: %v", err)
			}

			if success, ok := body["success"].(bool); ok && success != tt.wantSuccess {
				t.Errorf("Expected success=%v, got %v", tt.wantSuccess, success)
			}
		})
	}
}

func TestHandler_ValidatePostal(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		event      domain.LambdaEvent
		wantStatus int
		wantValid  bool
	}{
		{
			name: "valid postal code",
			event: domain.LambdaEvent{
				Body: `{"operation":"validate-postal","postalCode":"28001"}`,
			},
			wantStatus: 200,
			wantValid:  true,
		},
		{
			name: "invalid postal code",
			event: domain.LambdaEvent{
				Body: `{"operation":"validate-postal","postalCode":"99999"}`,
			},
			wantStatus: 200,
			wantValid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := Handler(ctx, tt.event)

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if response.StatusCode != tt.wantStatus {
				t.Errorf("Expected status code %d, got %d", tt.wantStatus, response.StatusCode)
			}

			var body domain.ValidationResult
			if err := json.Unmarshal([]byte(response.Body), &body); err != nil {
				t.Fatalf("Failed to parse response body: %v", err)
			}

			if body.Valid != tt.wantValid {
				t.Errorf("Expected valid=%v, got %v", tt.wantValid, body.Valid)
			}
		})
	}
}

func TestHandler_ValidateMunicipality(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		event      domain.LambdaEvent
		wantStatus int
		wantValid  bool
	}{
		{
			name: "valid municipality",
			event: domain.LambdaEvent{
				Body: `{"operation":"validate-municipality","municipality":"Madrid"}`,
			},
			wantStatus: 200,
			wantValid:  true,
		},
		{
			name: "invalid municipality",
			event: domain.LambdaEvent{
				Body: `{"operation":"validate-municipality","municipality":"NonExistentCity"}`,
			},
			wantStatus: 200,
			wantValid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := Handler(ctx, tt.event)

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if response.StatusCode != tt.wantStatus {
				t.Errorf("Expected status code %d, got %d", tt.wantStatus, response.StatusCode)
			}

			var body domain.ValidationResult
			if err := json.Unmarshal([]byte(response.Body), &body); err != nil {
				t.Fatalf("Failed to parse response body: %v", err)
			}

			if body.Valid != tt.wantValid {
				t.Errorf("Expected valid=%v, got %v", tt.wantValid, body.Valid)
			}
		})
	}
}

func TestHandler_AutocompletePostal(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		event      domain.LambdaEvent
		wantStatus int
		wantCount  int
	}{
		{
			name: "valid prefix",
			event: domain.LambdaEvent{
				Body: `{"operation":"autocomplete-postal","prefix":"280","limit":5}`,
			},
			wantStatus: 200,
			wantCount:  5,
		},
		{
			name: "missing prefix",
			event: domain.LambdaEvent{
				Body: `{"operation":"autocomplete-postal","limit":5}`,
			},
			wantStatus: 400,
			wantCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := Handler(ctx, tt.event)

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if response.StatusCode != tt.wantStatus {
				t.Errorf("Expected status code %d, got %d", tt.wantStatus, response.StatusCode)
			}

			if tt.wantStatus == 200 {
				var body map[string]interface{}
				if err := json.Unmarshal([]byte(response.Body), &body); err != nil {
					t.Fatalf("Failed to parse response body: %v", err)
				}

				if count, ok := body["count"].(float64); ok {
					if int(count) > tt.wantCount {
						t.Errorf("Expected at most %d results, got %d", tt.wantCount, int(count))
					}
				}
			}
		})
	}
}

func TestHandler_AutocompleteMunicipality(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		event      domain.LambdaEvent
		wantStatus int
		wantCount  int
	}{
		{
			name: "valid query",
			event: domain.LambdaEvent{
				Body: `{"operation":"autocomplete-municipality","query":"mad","limit":5}`,
			},
			wantStatus: 200,
			wantCount:  5,
		},
		{
			name: "missing query",
			event: domain.LambdaEvent{
				Body: `{"operation":"autocomplete-municipality","limit":5}`,
			},
			wantStatus: 400,
			wantCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := Handler(ctx, tt.event)

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if response.StatusCode != tt.wantStatus {
				t.Errorf("Expected status code %d, got %d", tt.wantStatus, response.StatusCode)
			}

			if tt.wantStatus == 200 {
				var body map[string]interface{}
				if err := json.Unmarshal([]byte(response.Body), &body); err != nil {
					t.Fatalf("Failed to parse response body: %v", err)
				}

				if count, ok := body["count"].(float64); ok {
					if int(count) > tt.wantCount {
						t.Errorf("Expected at most %d results, got %d", tt.wantCount, int(count))
					}
				}
			}
		})
	}
}

func TestHandler_InvalidOperation(t *testing.T) {
	ctx := context.Background()

	event := domain.LambdaEvent{
		Body: `{"operation":"invalid-operation","postalCode":"28001"}`,
	}

	response, err := Handler(ctx, event)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if response.StatusCode != 400 {
		t.Errorf("Expected status code 400, got %d", response.StatusCode)
	}

	var body map[string]interface{}
	if err := json.Unmarshal([]byte(response.Body), &body); err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}

	if success, ok := body["success"].(bool); ok && success {
		t.Errorf("Expected success=false, got %v", success)
	}
}

func TestHandler_InvalidJSON(t *testing.T) {
	ctx := context.Background()

	event := domain.LambdaEvent{
		Body: `invalid json`,
	}

	response, err := Handler(ctx, event)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if response.StatusCode != 400 {
		t.Errorf("Expected status code 400, got %d", response.StatusCode)
	}

	var body map[string]interface{}
	if err := json.Unmarshal([]byte(response.Body), &body); err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}

	if success, ok := body["success"].(bool); ok && success {
		t.Errorf("Expected success=false, got %v", success)
	}
}

func TestHandler_MissingOperation(t *testing.T) {
	ctx := context.Background()

	event := domain.LambdaEvent{
		Body: `{"postalCode":"28001"}`,
	}

	response, err := Handler(ctx, event)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if response.StatusCode != 400 {
		t.Errorf("Expected status code 400, got %d", response.StatusCode)
	}

	var body map[string]interface{}
	if err := json.Unmarshal([]byte(response.Body), &body); err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}

	if success, ok := body["success"].(bool); ok && success {
		t.Errorf("Expected success=false, got %v", success)
	}
}

// TestHandlerEdgeCases tests edge cases for better coverage
func TestHandlerEdgeCases(t *testing.T) {
	t.Run("invalid event format", func(t *testing.T) {
		// Create a complex struct that will fail JSON marshaling
		event := make(chan int) // channels cannot be marshaled to JSON
		response, err := Handler(context.Background(), event)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if response.StatusCode != 400 {
			t.Errorf("Expected status code 400, got %d", response.StatusCode)
		}
	})

	t.Run("API Gateway format with body", func(t *testing.T) {
		event := map[string]interface{}{
			"body": `{"operation":"validate-postal","postalCode":"28001"}`,
		}
		response, err := Handler(context.Background(), event)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if response.StatusCode != 200 {
			t.Errorf("Expected status code 200, got %d", response.StatusCode)
		}
	})

	t.Run("unknown operation", func(t *testing.T) {
		event := map[string]interface{}{
			"operation": "unknown-operation",
		}
		response, err := Handler(context.Background(), event)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if response.StatusCode != 400 {
			t.Errorf("Expected status code 400, got %d", response.StatusCode)
		}
	})
}
