// Package domain contains core business logic and domain models for the geocoding service.
package domain

// Constants for Portuguese postal code geocoding service.

// PostalCodeRegexPattern is the regex pattern for Portuguese postal codes.
// Format: 7 digits or 4 digits + hyphen + 3 digits (e.g., "1000205" or "1000-205").
const PostalCodeRegexPattern = `^\d{4}-\d{3}$|^\d{7}$`

// EarthRadiusKm is the Earth's radius in kilometers for Haversine distance calculations.
const EarthRadiusKm = 6371

// Coordinate validation ranges.
const (
	// MinLatitude is the minimum valid latitude (-90).
	MinLatitude = -90
	// MaxLatitude is the maximum valid latitude (90).
	MaxLatitude = 90
	// MinLongitude is the minimum valid longitude (-180).
	MinLongitude = -180
	// MaxLongitude is the maximum valid longitude (180).
	MaxLongitude = 180
)

// Default and maximum limits for autocomplete operations.
const (
	// DefaultAutocompleteLimit is the default number of results for autocomplete (10).
	DefaultAutocompleteLimit = 10
	// MaxAutocompleteLimit is the maximum number of results for autocomplete (50).
	MaxAutocompleteLimit = 50
)
