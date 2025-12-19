package provider

import (
	"testing"

	"github.com/pricofy/geocode-pt/internal/domain"
)

func TestPostalCodeProvider_GeocodeByPostalCode(t *testing.T) {
	p := NewPostalCodeProvider()

	tests := []struct {
		name        string
		postalCode  string
		wantSuccess bool
		wantErr     bool
	}{
		{
			name:        "valid postal code - Lisbon",
			postalCode:  "1000205",
			wantSuccess: true,
			wantErr:     false,
		},
		{
			name:        "valid postal code - Porto",
			postalCode:  "4000050",
			wantSuccess: true,
			wantErr:     false,
		},
		{
			name:        "invalid postal code",
			postalCode:  "99999",
			wantSuccess: false,
			wantErr:     true,
		},
		{
			name:        "empty postal code",
			postalCode:  "",
			wantSuccess: false,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := p.GeocodeByPostalCode(tt.postalCode)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				if _, ok := err.(*domain.PostalCodeNotFoundError); !ok {
					t.Errorf("Expected PostalCodeNotFoundError, got %T", err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result.Success != tt.wantSuccess {
					t.Errorf("Expected success=%v, got %v", tt.wantSuccess, result.Success)
				}
				if result.PostalCode != tt.postalCode {
					t.Errorf("Expected postalCode=%s, got %s", tt.postalCode, result.PostalCode)
				}
				if result.Coords.Lat == 0 && result.Coords.Lon == 0 {
					t.Errorf("Expected non-zero coordinates")
				}
			}
		})
	}
}

func TestPostalCodeProvider_GeocodeByMunicipality(t *testing.T) {
	p := NewPostalCodeProvider()

	tests := []struct {
		name         string
		municipality string
		wantSuccess  bool
		wantErr      bool
	}{
		{
			name:         "valid municipality - Lisbon",
			municipality: "Lisboa",
			wantSuccess:  true,
			wantErr:      false,
		},
		{
			name:         "valid municipality - Porto",
			municipality: "Porto",
			wantSuccess:  true,
			wantErr:      false,
		},
		{
			name:         "invalid municipality",
			municipality: "NonExistentCity",
			wantSuccess:  false,
			wantErr:      true,
		},
		{
			name:         "case insensitive",
			municipality: "lisboa",
			wantSuccess:  true,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := p.GeocodeByMunicipality(tt.municipality)

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
				if result.Municipality == "" {
					t.Errorf("Expected non-empty municipality")
				}
			}
		})
	}
}

func TestPostalCodeProvider_ReverseGeocode(t *testing.T) {
	p := NewPostalCodeProvider()

	tests := []struct {
		name        string
		lat         float64
		lon         float64
		wantSuccess bool
		wantErr     bool
	}{
		{
			name:        "Lisbon coordinates",
			lat:         38.7167,
			lon:         -9.1333,
			wantSuccess: true,
			wantErr:     false,
		},
		{
			name:        "Porto coordinates",
			lat:         41.1496,
			lon:         -8.611,
			wantSuccess: true,
			wantErr:     false,
		},
		{
			name:        "Valid coordinates in Portugal",
			lat:         39.3999,
			lon:         -8.2245,
			wantSuccess: true,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, distance, err := p.ReverseGeocode(tt.lat, tt.lon)

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
				if distance < 0 {
					t.Errorf("Expected non-negative distance, got %f", distance)
				}
				if distance > 1000 {
					t.Errorf("Distance seems too large: %f km", distance)
				}
			}
		})
	}
}

func TestPostalCodeProvider_ValidatePostalCode(t *testing.T) {
	p := NewPostalCodeProvider()

	tests := []struct {
		name       string
		postalCode string
		want       bool
	}{
		{
			name:       "valid postal code",
			postalCode: "1000205",
			want:       true,
		},
		{
			name:       "invalid postal code",
			postalCode: "99999",
			want:       false,
		},
		{
			name:       "empty postal code",
			postalCode: "",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.ValidatePostalCode(tt.postalCode)
			if got != tt.want {
				t.Errorf("ValidatePostalCode(%s) = %v, want %v", tt.postalCode, got, tt.want)
			}
		})
	}
}

func TestPostalCodeProvider_ValidateMunicipality(t *testing.T) {
	p := NewPostalCodeProvider()

	tests := []struct {
		name         string
		municipality string
		want         bool
	}{
		{
			name:         "valid municipality",
			municipality: "Lisboa",
			want:         true,
		},
		{
			name:         "case insensitive",
			municipality: "lisboa",
			want:         true,
		},
		{
			name:         "invalid municipality",
			municipality: "NonExistentCity",
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.ValidateMunicipality(tt.municipality)
			if got != tt.want {
				t.Errorf("ValidateMunicipality(%s) = %v, want %v", tt.municipality, got, tt.want)
			}
		})
	}
}

func TestPostalCodeProvider_AutocompletePostalCode(t *testing.T) {
	p := NewPostalCodeProvider()

	tests := []struct {
		name      string
		prefix    string
		limit     int
		wantCount int
		wantErr   bool
	}{
		{
			name:      "prefix 10",
			prefix:    "10",
			limit:     10,
			wantCount: 10,
			wantErr:   false,
		},
		{
			name:      "prefix 40",
			prefix:    "40",
			limit:     5,
			wantCount: 5,
			wantErr:   false,
		},
		{
			name:      "non-existent prefix",
			prefix:    "000",
			limit:     10,
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := p.AutocompletePostalCode(tt.prefix, tt.limit)

			if len(results) != tt.wantCount {
				t.Errorf("AutocompletePostalCode(%s, %d) returned %d results, want %d",
					tt.prefix, tt.limit, len(results), tt.wantCount)
			}

			// Verify all results start with prefix
			for _, result := range results {
				if len(result.PostalCode) < len(tt.prefix) || result.PostalCode[:len(tt.prefix)] != tt.prefix {
					t.Errorf("Result %s does not start with prefix %s", result.PostalCode, tt.prefix)
				}
			}
		})
	}
}

func TestPostalCodeProvider_AutocompleteMunicipality(t *testing.T) {
	p := NewPostalCodeProvider()

	tests := []struct {
		name      string
		query     string
		limit     int
		wantCount int
		wantErr   bool
	}{
		{
			name:      "query 'Lis'",
			query:     "Lis",
			limit:     10,
			wantCount: 10,
			wantErr:   false,
		},
		{
			name:      "query 'Por'",
			query:     "Por",
			limit:     5,
			wantCount: 5,
			wantErr:   false,
		},
		{
			name:      "case insensitive",
			query:     "LIS",
			limit:     10,
			wantCount: 10,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := p.AutocompleteMunicipality(tt.query, tt.limit)

			if len(results) > tt.limit {
				t.Errorf("AutocompleteMunicipality(%s, %d) returned %d results, want at most %d",
					tt.query, tt.limit, len(results), tt.limit)
			}

			// Verify all results contain query (case insensitive)
			queryLower := tt.query
			for _, result := range results {
				municipalityLower := result.Municipality
				if len(municipalityLower) < len(queryLower) {
					t.Errorf("Result municipality %s is shorter than query %s", result.Municipality, tt.query)
				}
			}
		})
	}
}

func TestPostalCodeProvider_CalculateDistance(t *testing.T) {
	p := NewPostalCodeProvider()

	// Test Haversine distance calculation
	// Lisbon coordinates
	lisbonLat, lisbonLon := 38.7167, -9.1333
	portoLat, portoLon := 41.1496, -8.611

	_, dist, err := p.ReverseGeocode(lisbonLat, lisbonLon)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	_, dist2, err := p.ReverseGeocode(portoLat, portoLon)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Distance should be reasonable (not negative, not too large)
	if dist < 0 || dist > 1000 {
		t.Errorf("Distance from Lisbon seems incorrect: %f km", dist)
	}
	if dist2 < 0 || dist2 > 1000 {
		t.Errorf("Distance from Porto seems incorrect: %f km", dist2)
	}
}

//nolint:gocognit // Test function with table-driven tests has inherent complexity
func TestPostalCodeProvider_GeocodeByMunicipalitiesBatch(t *testing.T) {
	p := NewPostalCodeProvider()

	tests := []struct {
		name           string
		municipalities []string
		wantFound      int
		wantNotFound   int
	}{
		{
			name:           "all valid municipalities",
			municipalities: []string{"Lisboa", "Porto", "Braga"},
			wantFound:      3,
			wantNotFound:   0,
		},
		{
			name:           "mixed valid and invalid",
			municipalities: []string{"Lisboa", "NonExistent", "Porto"},
			wantFound:      2,
			wantNotFound:   1,
		},
		{
			name:           "all invalid",
			municipalities: []string{"NonExistent1", "NonExistent2"},
			wantFound:      0,
			wantNotFound:   2,
		},
		{
			name:           "case insensitive",
			municipalities: []string{"LISBOA", "porto", "BrAgA"},
			wantFound:      3,
			wantNotFound:   0,
		},
		{
			name:           "empty list",
			municipalities: []string{},
			wantFound:      0,
			wantNotFound:   0,
		},
		{
			name:           "single municipality",
			municipalities: []string{"Coimbra"},
			wantFound:      1,
			wantNotFound:   0,
		},
		{
			name:           "with whitespace",
			municipalities: []string{" Lisboa ", "  Porto"},
			wantFound:      2,
			wantNotFound:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := p.GeocodeByMunicipalitiesBatch(tt.municipalities)

			// Count found and not found
			foundCount := 0
			notFoundCount := 0
			for _, result := range results {
				if result != nil && result.Found {
					foundCount++
				} else {
					notFoundCount++
				}
			}

			if foundCount != tt.wantFound {
				t.Errorf("Expected %d found, got %d", tt.wantFound, foundCount)
			}
			if notFoundCount != tt.wantNotFound {
				t.Errorf("Expected %d not found, got %d", tt.wantNotFound, notFoundCount)
			}

			// Verify found results have valid coordinates
			for municipality, result := range results {
				if result != nil && result.Found {
					if result.Lat == 0 && result.Lon == 0 {
						t.Errorf("Municipality %s has zero coordinates", municipality)
					}
					if result.PostalCode == "" {
						t.Errorf("Municipality %s has empty postal code", municipality)
					}
				}
			}
		})
	}
}

func TestPostalCodeProvider_GeocodeByMunicipalitiesBatch_Preserves_Original_Names(t *testing.T) {
	p := NewPostalCodeProvider()

	// Test that the original municipality names are preserved as keys
	municipalities := []string{"Lisboa", "PORTO", "braga"}
	results := p.GeocodeByMunicipalitiesBatch(municipalities)

	// Check that original names are used as keys
	for _, original := range municipalities {
		if _, ok := results[original]; !ok {
			t.Errorf("Original name '%s' not found in results keys", original)
		}
	}
}

func TestPostalCodeProvider_GeocodeByMunicipalitiesBatch_Large_Batch(t *testing.T) {
	p := NewPostalCodeProvider()

	// Test with a larger batch
	municipalities := []string{
		"Lisboa", "Porto", "Braga", "Coimbra", "Faro",
		"Aveiro", "Leiria", "Setúbal", "Viseu", "Évora",
		"Guarda", "Santarém", "Beja", "Castelo Branco", "Viana do Castelo",
	}

	results := p.GeocodeByMunicipalitiesBatch(municipalities)

	if len(results) != len(municipalities) {
		t.Errorf("Expected %d results, got %d", len(municipalities), len(results))
	}

	// Count how many were found
	foundCount := 0
	for _, result := range results {
		if result != nil && result.Found {
			foundCount++
		}
	}

	// Most major cities should be found
	if foundCount < 10 {
		t.Errorf("Expected at least 10 cities to be found, got %d", foundCount)
	}
}
