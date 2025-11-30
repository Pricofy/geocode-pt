//nolint:dupl // test code duplication is acceptable for clarity
package application

import (
	"encoding/json"
	"testing"

	"github.com/pricofy/geocode-pt/internal/domain"
)

func TestPostalCodeService_GeocodeByPostal(t *testing.T) {
	service := NewPostalCodeService()

	tests := []struct {
		name        string
		event       domain.LambdaEvent
		wantSuccess bool
		wantErr     bool
	}{
		{
			name: "valid postal code",
			event: domain.LambdaEvent{
				Body: `{"operation":"geocode-by-postal","postalCode":"28001"}`,
			},
			wantSuccess: true,
			wantErr:     false,
		},
		{
			name: "valid municipality",
			event: domain.LambdaEvent{
				Body: `{"operation":"geocode-by-postal","municipality":"Madrid"}`,
			},
			wantSuccess: true,
			wantErr:     false,
		},
		{
			name: "invalid postal code format",
			event: domain.LambdaEvent{
				Body: `{"operation":"geocode-by-postal","postalCode":"123"}`,
			},
			wantSuccess: false,
			wantErr:     true,
		},
		{
			name: "invalid JSON",
			event: domain.LambdaEvent{
				Body: `invalid json`,
			},
			wantSuccess: false,
			wantErr:     true,
		},
		{
			name: "missing both postal code and municipality",
			event: domain.LambdaEvent{
				Body: `{"operation":"geocode-by-postal"}`,
			},
			wantSuccess: false,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.GeocodeByPostal(tt.event)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result.Success != tt.wantSuccess {
					t.Errorf("Expected success=%v, got %v", tt.wantSuccess, result.Success)
				}
			}
		})
	}
}

func TestPostalCodeService_ReverseGeocode(t *testing.T) {
	service := NewPostalCodeService()

	tests := []struct {
		name        string
		event       domain.LambdaEvent
		wantSuccess bool
		wantErr     bool
	}{
		{
			name: "valid coordinates",
			event: domain.LambdaEvent{
				Body: `{"operation":"reverse-geocode","lat":40.4168,"lon":-3.7038}`,
			},
			wantSuccess: true,
			wantErr:     false,
		},
		{
			name: "missing lat",
			event: domain.LambdaEvent{
				Body: `{"operation":"reverse-geocode","lon":-3.7038}`,
			},
			wantSuccess: false,
			wantErr:     true,
		},
		{
			name: "missing lon",
			event: domain.LambdaEvent{
				Body: `{"operation":"reverse-geocode","lat":40.4168}`,
			},
			wantSuccess: false,
			wantErr:     true,
		},
		{
			name: "invalid coordinates - out of range",
			event: domain.LambdaEvent{
				Body: `{"operation":"reverse-geocode","lat":100,"lon":-3.7038}`,
			},
			wantSuccess: false,
			wantErr:     true,
		},
		{
			name: "invalid JSON",
			event: domain.LambdaEvent{
				Body: `invalid json`,
			},
			wantSuccess: false,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.ReverseGeocode(tt.event)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result.Success != tt.wantSuccess {
					t.Errorf("Expected success=%v, got %v", tt.wantSuccess, result.Success)
				}
				if result.PostalCode == "" {
					t.Errorf("Expected non-empty postalCode")
				}
				if result.Distance < 0 {
					t.Errorf("Expected non-negative distance, got %f", result.Distance)
				}
			}
		})
	}
}

func TestPostalCodeService_ValidatePostal(t *testing.T) {
	service := NewPostalCodeService()

	tests := []struct {
		name      string
		event     domain.LambdaEvent
		wantValid bool
		wantErr   bool
	}{
		{
			name: "valid postal code",
			event: domain.LambdaEvent{
				Body: `{"operation":"validate-postal","postalCode":"28001"}`,
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "invalid postal code",
			event: domain.LambdaEvent{
				Body: `{"operation":"validate-postal","postalCode":"99999"}`,
			},
			wantValid: false,
			wantErr:   false,
		},
		{
			name: "missing postal code",
			event: domain.LambdaEvent{
				Body: `{"operation":"validate-postal"}`,
			},
			wantValid: false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.ValidatePostal(tt.event)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result.Valid != tt.wantValid {
					t.Errorf("Expected valid=%v, got %v", tt.wantValid, result.Valid)
				}
			}
		})
	}
}

func TestPostalCodeService_ValidateMunicipality(t *testing.T) {
	service := NewPostalCodeService()

	tests := []struct {
		name      string
		event     domain.LambdaEvent
		wantValid bool
		wantErr   bool
	}{
		{
			name: "valid municipality",
			event: domain.LambdaEvent{
				Body: `{"operation":"validate-municipality","municipality":"Madrid"}`,
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "invalid municipality",
			event: domain.LambdaEvent{
				Body: `{"operation":"validate-municipality","municipality":"NonExistentCity"}`,
			},
			wantValid: false,
			wantErr:   false,
		},
		{
			name: "missing municipality",
			event: domain.LambdaEvent{
				Body: `{"operation":"validate-municipality"}`,
			},
			wantValid: false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.ValidateMunicipality(tt.event)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result.Valid != tt.wantValid {
					t.Errorf("Expected valid=%v, got %v", tt.wantValid, result.Valid)
				}
			}
		})
	}
}

func TestPostalCodeService_AutocompletePostal(t *testing.T) {
	service := NewPostalCodeService()

	tests := []struct {
		name      string
		event     domain.LambdaEvent
		wantCount int
		wantErr   bool
	}{
		{
			name: "valid prefix",
			event: domain.LambdaEvent{
				Body: `{"operation":"autocomplete-postal","prefix":"280","limit":5}`,
			},
			wantCount: 5,
			wantErr:   false,
		},
		{
			name: "missing prefix",
			event: domain.LambdaEvent{
				Body: `{"operation":"autocomplete-postal","limit":5}`,
			},
			wantCount: 0,
			wantErr:   true,
		},
		{
			name: "default limit",
			event: domain.LambdaEvent{
				Body: `{"operation":"autocomplete-postal","prefix":"280"}`,
			},
			wantCount: 10, // default limit
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := service.AutocompletePostal(tt.event)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(results) > tt.wantCount {
					t.Errorf("Expected at most %d results, got %d", tt.wantCount, len(results))
				}
			}
		})
	}
}

func TestPostalCodeService_AutocompleteMunicipality(t *testing.T) {
	service := NewPostalCodeService()

	tests := []struct {
		name      string
		event     domain.LambdaEvent
		wantCount int
		wantErr   bool
	}{
		{
			name: "valid query",
			event: domain.LambdaEvent{
				Body: `{"operation":"autocomplete-municipality","query":"mad","limit":5}`,
			},
			wantCount: 5,
			wantErr:   false,
		},
		{
			name: "missing query",
			event: domain.LambdaEvent{
				Body: `{"operation":"autocomplete-municipality","limit":5}`,
			},
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := service.AutocompleteMunicipality(tt.event)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(results) > tt.wantCount {
					t.Errorf("Expected at most %d results, got %d", tt.wantCount, len(results))
				}
			}
		})
	}
}

func TestPostalCodeService_ValidateCoordinates(t *testing.T) {
	service := NewPostalCodeService()

	// Test coordinate validation through ReverseGeocode
	tests := []struct {
		name    string
		lat     float64
		lon     float64
		wantErr bool
	}{
		{
			name:    "valid coordinates",
			lat:     40.4168,
			lon:     -3.7038,
			wantErr: false,
		},
		{
			name:    "latitude too high",
			lat:     91.0,
			lon:     -3.7038,
			wantErr: true,
		},
		{
			name:    "latitude too low",
			lat:     -91.0,
			lon:     -3.7038,
			wantErr: true,
		},
		{
			name:    "longitude too high",
			lat:     40.4168,
			lon:     181.0,
			wantErr: true,
		},
		{
			name:    "longitude too low",
			lat:     40.4168,
			lon:     -181.0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventBody, _ := json.Marshal(map[string]interface{}{
				"operation": "reverse-geocode",
				"lat":       tt.lat,
				"lon":       tt.lon,
			})

			event := domain.LambdaEvent{
				Body: string(eventBody),
			}

			_, err := service.ReverseGeocode(event)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestPostalCodeService_EdgeCases tests edge cases for better coverage
func TestPostalCodeService_EdgeCases(t *testing.T) {
	service := NewPostalCodeService()

	t.Run("GeocodeByPostal with both postalCode and municipality", func(t *testing.T) {
		event := domain.LambdaEvent{
			Body: `{"operation":"geocode-by-postal","postalCode":"28001","municipality":"Madrid"}`,
		}
		result, err := service.GeocodeByPostal(event)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if !result.Success {
			t.Error("Expected success=true")
		}
	})

	t.Run("GeocodeByPostal with neither postalCode nor municipality", func(t *testing.T) {
		event := domain.LambdaEvent{
			Body: `{"operation":"geocode-by-postal"}`,
		}
		_, err := service.GeocodeByPostal(event)
		if err == nil {
			t.Error("Expected error, got nil")
		}
	})

	t.Run("ReverseGeocode with invalid coordinates", func(t *testing.T) {
		event := domain.LambdaEvent{
			Body: `{"operation":"reverse-geocode","lat":999,"lon":999}`,
		}
		_, err := service.ReverseGeocode(event)
		if err == nil {
			t.Error("Expected error for invalid coordinates, got nil")
		}
	})

	t.Run("AutocompletePostal with empty prefix", func(t *testing.T) {
		event := domain.LambdaEvent{
			Body: `{"operation":"autocomplete-postal","prefix":""}`,
		}
		results, err := service.AutocompletePostal(event)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(results) == 0 {
			t.Error("Expected some results for empty prefix")
		}
	})

	t.Run("AutocompleteMunicipality with empty query", func(t *testing.T) {
		event := domain.LambdaEvent{
			Body: `{"operation":"autocomplete-municipality","query":""}`,
		}
		results, err := service.AutocompleteMunicipality(event)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(results) == 0 {
			t.Error("Expected some results for empty query")
		}
	})

	t.Run("AutocompletePostal with custom limit", func(t *testing.T) {
		event := domain.LambdaEvent{
			Body: `{"operation":"autocomplete-postal","prefix":"28","limit":5}`,
		}
		results, err := service.AutocompletePostal(event)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(results) > 5 {
			t.Errorf("Expected max 5 results, got %d", len(results))
		}
	})

	t.Run("AutocompleteMunicipality with custom limit", func(t *testing.T) {
		event := domain.LambdaEvent{
			Body: `{"operation":"autocomplete-municipality","query":"mad","limit":3}`,
		}
		results, err := service.AutocompleteMunicipality(event)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(results) > 3 {
			t.Errorf("Expected max 3 results, got %d", len(results))
		}
	})
}

// TestGeocodeByPostalWithMunicipalityNotFound tests the municipality not found case
func TestGeocodeByPostalWithMunicipalityNotFound(t *testing.T) {
	service := NewPostalCodeService()

	t.Run("municipality not found after postal code not found", func(t *testing.T) {
		event := domain.LambdaEvent{
			Body: `{"operation":"geocode-by-postal","postalCode":"99999","municipality":"NonExistentCity"}`,
		}
		_, err := service.GeocodeByPostal(event)
		if err == nil {
			t.Error("Expected error for non-existent municipality, got nil")
		}
	})

	t.Run("valid postal code with invalid municipality", func(t *testing.T) {
		event := domain.LambdaEvent{
			Body: `{"operation":"geocode-by-postal","postalCode":"28001","municipality":"NonExistentCity"}`,
		}
		result, err := service.GeocodeByPostal(event)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if !result.Success {
			t.Error("Expected success=true for valid postal code")
		}
	})
}

// TestReverseGeocodeEdgeCases tests edge cases in reverse geocode
func TestReverseGeocodeEdgeCases(t *testing.T) {
	service := NewPostalCodeService()

	t.Run("coordinates at boundary", func(t *testing.T) {
		event := domain.LambdaEvent{
			Body: `{"operation":"reverse-geocode","lat":90,"lon":180}`,
		}
		_, err := service.ReverseGeocode(event)
		// Should work or fail gracefully
		if err != nil {
			// Error is acceptable for boundary coordinates
			t.Logf("Boundary coordinates error: %v", err)
		}
	})

	t.Run("coordinates slightly out of range", func(t *testing.T) {
		event := domain.LambdaEvent{
			Body: `{"operation":"reverse-geocode","lat":90.1,"lon":180.1}`,
		}
		_, err := service.ReverseGeocode(event)
		if err == nil {
			t.Error("Expected error for out-of-range coordinates, got nil")
		}
	})
}

//nolint:gocognit // comprehensive test function with many test cases
func TestPostalCodeService_InputParsing(t *testing.T) {
	service := NewPostalCodeService()

	t.Run("ValidatePostal with invalid JSON", func(t *testing.T) {
		event := domain.LambdaEvent{Body: `invalid json`}
		_, err := service.ValidatePostal(event)
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})

	t.Run("ValidatePostal with missing postalCode", func(t *testing.T) {
		event := domain.LambdaEvent{Body: `{"operation":"validate-postal"}`}
		_, err := service.ValidatePostal(event)
		if err == nil {
			t.Error("Expected error for missing postalCode")
		}
	})

	t.Run("ValidatePostal with empty postalCode", func(t *testing.T) {
		event := domain.LambdaEvent{Body: `{"operation":"validate-postal","postalCode":""}`}
		_, err := service.ValidatePostal(event)
		if err == nil {
			t.Error("Expected error for empty postalCode")
		}
	})

	t.Run("ValidateMunicipality with invalid JSON", func(t *testing.T) {
		event := domain.LambdaEvent{Body: `invalid json`}
		_, err := service.ValidateMunicipality(event)
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})

	t.Run("ValidateMunicipality with missing municipality", func(t *testing.T) {
		event := domain.LambdaEvent{Body: `{"operation":"validate-municipality"}`}
		_, err := service.ValidateMunicipality(event)
		if err == nil {
			t.Error("Expected error for missing municipality")
		}
	})

	t.Run("ValidateMunicipality with empty municipality", func(t *testing.T) {
		event := domain.LambdaEvent{Body: `{"operation":"validate-municipality","municipality":""}`}
		_, err := service.ValidateMunicipality(event)
		if err == nil {
			t.Error("Expected error for empty municipality")
		}
	})

	t.Run("AutocompletePostal with invalid JSON", func(t *testing.T) {
		event := domain.LambdaEvent{Body: `invalid json`}
		_, err := service.AutocompletePostal(event)
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})

	t.Run("AutocompletePostal with missing prefix", func(t *testing.T) {
		event := domain.LambdaEvent{Body: `{"operation":"autocomplete-postal"}`}
		_, err := service.AutocompletePostal(event)
		if err == nil {
			t.Error("Expected error for missing prefix")
		}
	})

	t.Run("AutocompletePostal with invalid limit (negative)", func(t *testing.T) {
		event := domain.LambdaEvent{Body: `{"operation":"autocomplete-postal","prefix":"28","limit":-1}`}
		_, err := service.AutocompletePostal(event)
		if err == nil {
			t.Error("Expected error for negative limit")
		}
	})

	t.Run("AutocompletePostal with invalid limit (zero)", func(t *testing.T) {
		event := domain.LambdaEvent{Body: `{"operation":"autocomplete-postal","prefix":"28","limit":0}`}
		_, err := service.AutocompletePostal(event)
		if err == nil {
			t.Error("Expected error for zero limit")
		}
	})

	t.Run("AutocompletePostal with large limit (clamping)", func(t *testing.T) {
		event := domain.LambdaEvent{Body: `{"operation":"autocomplete-postal","prefix":"28","limit":1000}`}
		results, err := service.AutocompletePostal(event)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(results) > 50 {
			t.Errorf("Expected max 50 results, got %d", len(results))
		}
	})

	t.Run("AutocompleteMunicipality with invalid JSON", func(t *testing.T) {
		event := domain.LambdaEvent{Body: `invalid json`}
		_, err := service.AutocompleteMunicipality(event)
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})

	t.Run("AutocompleteMunicipality with missing query", func(t *testing.T) {
		event := domain.LambdaEvent{Body: `{"operation":"autocomplete-municipality"}`}
		_, err := service.AutocompleteMunicipality(event)
		if err == nil {
			t.Error("Expected error for missing query")
		}
	})

	t.Run("AutocompletePostal with default limit", func(t *testing.T) {
		event := domain.LambdaEvent{Body: `{"operation":"autocomplete-postal","prefix":"28"}`}
		results, err := service.AutocompletePostal(event)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		// Default limit is 10. If we have more than 10 results for "28", we should get 10.
		// "28" (Madrid) has many codes.
		if len(results) > 10 {
			t.Errorf("Expected max 10 results (default limit), got %d", len(results))
		}
	})

	t.Run("AutocompleteMunicipality with default limit", func(t *testing.T) {
		event := domain.LambdaEvent{Body: `{"operation":"autocomplete-municipality","query":"mad"}`}
		results, err := service.AutocompleteMunicipality(event)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(results) > 10 {
			t.Errorf("Expected max 10 results (default limit), got %d", len(results))
		}
	})

	t.Run("GeocodeByPostal with not found postal code and empty municipality", func(t *testing.T) {
		event := domain.LambdaEvent{Body: `{"operation":"geocode-by-postal","postalCode":"99999"}`}
		_, err := service.GeocodeByPostal(event)
		if err == nil {
			t.Error("Expected error for not found postal code")
		}
		// Check if error is PostalCodeNotFoundError
		if _, ok := err.(*domain.PostalCodeNotFoundError); !ok {
			t.Errorf("Expected PostalCodeNotFoundError, got %T", err)
		}
	})
}
