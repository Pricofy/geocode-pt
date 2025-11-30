package main

import (
	"context"
	"strings"
	"testing"

	"github.com/pricofy/geocode-pt/internal/domain"
)

func TestHandleRequest(t *testing.T) {
	app, err := NewApp()
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	ctx := context.Background()

	tests := []struct {
		name           string
		event          domain.LambdaEvent
		wantStatusCode int
		wantBody       string // Optional: check for substring in body
	}{
		// --- Geocode By Postal ---
		{
			name: "geocode-by-postal success",
			event: domain.LambdaEvent{
				Body: `{"operation":"geocode-by-postal","postalCode":"28001"}`,
			},
			wantStatusCode: 200,
			wantBody:       `"postalCode":"28001"`,
		},
		{
			name: "geocode-by-postal not found",
			event: domain.LambdaEvent{
				Body: `{"operation":"geocode-by-postal","postalCode":"99999"}`,
			},
			wantStatusCode: 404,
			wantBody:       "not found",
		},
		{
			name: "geocode-by-postal invalid format",
			event: domain.LambdaEvent{
				Body: `{"operation":"geocode-by-postal","postalCode":"123"}`,
			},
			wantStatusCode: 400,
			wantBody:       "Invalid postal code format",
		},

		// --- Reverse Geocode ---
		{
			name: "reverse-geocode success",
			event: domain.LambdaEvent{
				Body: `{"operation":"reverse-geocode","lat":40.4168,"lon":-3.7038}`,
			},
			wantStatusCode: 200,
			wantBody:       `"postalCode"`,
		},
		{
			name: "reverse-geocode invalid coordinates",
			event: domain.LambdaEvent{
				Body: `{"operation":"reverse-geocode","lat":999,"lon":999}`,
			},
			wantStatusCode: 400,
			wantBody:       "Invalid coordinates",
		},

		// --- Validate Postal ---
		{
			name: "validate-postal success",
			event: domain.LambdaEvent{
				Body: `{"operation":"validate-postal","postalCode":"28001"}`,
			},
			wantStatusCode: 200,
			wantBody:       `"valid":true`,
		},
		{
			name: "validate-postal invalid",
			event: domain.LambdaEvent{
				Body: `{"operation":"validate-postal","postalCode":"99999"}`,
			},
			wantStatusCode: 200,
			wantBody:       `"valid":false`,
		},
		{
			name: "validate-postal missing input",
			event: domain.LambdaEvent{
				Body: `{"operation":"validate-postal"}`,
			},
			wantStatusCode: 400,
			wantBody:       "required",
		},

		// --- Validate Municipality ---
		{
			name: "validate-municipality success",
			event: domain.LambdaEvent{
				Body: `{"operation":"validate-municipality","municipality":"Madrid"}`,
			},
			wantStatusCode: 200,
			wantBody:       `"valid":true`,
		},
		{
			name: "validate-municipality invalid",
			event: domain.LambdaEvent{
				Body: `{"operation":"validate-municipality","municipality":"NonExistent"}`,
			},
			wantStatusCode: 200,
			wantBody:       `"valid":false`,
		},
		{
			name: "validate-municipality missing input",
			event: domain.LambdaEvent{
				Body: `{"operation":"validate-municipality"}`,
			},
			wantStatusCode: 400,
			wantBody:       "required",
		},

		// --- Autocomplete Postal ---
		{
			name: "autocomplete-postal success",
			event: domain.LambdaEvent{
				Body: `{"operation":"autocomplete-postal","prefix":"280"}`,
			},
			wantStatusCode: 200,
			wantBody:       `"results"`,
		},
		{
			name: "autocomplete-postal missing prefix",
			event: domain.LambdaEvent{
				Body: `{"operation":"autocomplete-postal"}`,
			},
			wantStatusCode: 400,
			wantBody:       "required",
		},

		// --- Autocomplete Municipality ---
		{
			name: "autocomplete-municipality success",
			event: domain.LambdaEvent{
				Body: `{"operation":"autocomplete-municipality","query":"mad"}`,
			},
			wantStatusCode: 200,
			wantBody:       `"results"`,
		},
		{
			name: "autocomplete-municipality missing query",
			event: domain.LambdaEvent{
				Body: `{"operation":"autocomplete-municipality"}`,
			},
			wantStatusCode: 400,
			wantBody:       "required",
		},

		// --- General Errors ---
		{
			name: "invalid json body",
			event: domain.LambdaEvent{
				Body: `invalid json`,
			},
			wantStatusCode: 400,
			wantBody:       "Invalid JSON",
		},
		{
			name: "unknown operation",
			event: domain.LambdaEvent{
				Body: `{"operation":"unknown-op"}`,
			},
			wantStatusCode: 400,
			wantBody:       "Unknown operation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := app.HandleRequest(ctx, tt.event)

			// HandleRequest should not return error for these cases, but return error response
			if err != nil {
				t.Errorf("HandleRequest returned unexpected error: %v", err)
			}

			if response.StatusCode != tt.wantStatusCode {
				t.Errorf("Expected status code %d, got %d. Body: %s", tt.wantStatusCode, response.StatusCode, response.Body)
			}

			if tt.wantBody != "" {
				if !strings.Contains(response.Body, tt.wantBody) {
					t.Errorf("Expected body to contain %q, got %q", tt.wantBody, response.Body)
				}
			}
		})
	}
}

func TestInit(t *testing.T) {
	// Test that NewApp works (it's called in init but we can't easily test init directly without side effects)
	app, err := NewApp()
	if err != nil {
		t.Errorf("NewApp() error = %v", err)
	}
	//nolint:staticcheck // SA5011: intentional nil checks for test validation
	if app == nil {
		t.Error("NewApp() returned nil app")
	}
	//nolint:staticcheck // SA5011: intentional nil checks for test validation
	if app.service == nil {
		t.Error("NewApp() returned app with nil service")
	}
	//nolint:staticcheck // SA5011: intentional nil checks for test validation
	if app.logger == nil {
		t.Error("NewApp() returned app with nil logger")
	}
}
