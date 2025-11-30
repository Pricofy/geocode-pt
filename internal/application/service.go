package application

import (
	"encoding/json"
	"math"
	"regexp"

	"github.com/pricofy/geocode-pt/internal/domain"
	"github.com/pricofy/geocode-pt/internal/infrastructure/provider"
	"github.com/pricofy/geocode-pt/internal/shared/logger"
)

// serviceLogger is the logger instance for the service
var serviceLogger = logger.NewLogger("PostalCodeService")

// Error messages constants to avoid duplication
const (
	errorMessageFailedToParseRequestBody = "Failed to parse request body"
	errorMessageInvalidJSONInRequestBody = "Invalid JSON in request body"
)

// PostalCodeService provides business logic for Spanish postal code operations.
// Coordinates between the handler layer and the provider layer.
// Handles input validation, parsing, and error handling.
type PostalCodeService struct {
	provider *provider.PostalCodeProvider
}

// NewPostalCodeService creates a new PostalCodeService instance.
func NewPostalCodeService() *PostalCodeService {
	return &PostalCodeService{
		provider: provider.NewPostalCodeProvider(),
	}
}

// postalCodeRegex is the compiled regex for validating Spanish postal codes (5 digits).
var postalCodeRegex = regexp.MustCompile(domain.PostalCodeRegexPattern)

// validateCoordinates validates that coordinates are within valid ranges.
func (s *PostalCodeService) validateCoordinates(lat, lon float64) error {
	if math.IsNaN(lat) || math.IsNaN(lon) ||
		lat < domain.MinLatitude || lat > domain.MaxLatitude ||
		lon < domain.MinLongitude || lon > domain.MaxLongitude {
		return domain.NewInvalidCoordinatesError("Invalid coordinates", lat, lon)
	}
	return nil
}

// GeocodeByPostal geocodes a postal code or municipality to coordinates.
// Tries postal code first (faster, O(1)), then municipality if not found (O(n)).
func (s *PostalCodeService) GeocodeByPostal(event domain.LambdaEvent) (domain.GeocodingResult, error) {
	postalCode, municipality, err := s.parseGeocodingInput(event)
	if err != nil {
		return domain.GeocodingResult{}, err
	}

	serviceLogger.Debug("Geocoding by postal", map[string]interface{}{
		"postalCode":   postalCode,
		"municipality": municipality,
	})

	// Try postal code first (most precise, O(1))
	if postalCode != "" {
		result, err := s.tryGeocodeByPostalCode(postalCode, municipality)
		if err == nil {
			return result, nil
		}
		// If postal code not found and we have municipality, continue to try municipality
		if _, ok := err.(*domain.PostalCodeNotFoundError); !ok || municipality == "" {
			return domain.GeocodingResult{}, err
		}
	}

	// Try municipality (less precise, O(n))
	if municipality != "" {
		return s.geocodeByMunicipality(municipality)
	}

	// Neither postal code nor municipality provided
	return domain.GeocodingResult{}, domain.NewPostalCodeNotFoundError("", "")
}

// parseGeocodingInput parses and validates geocoding input from Lambda event.
func (s *PostalCodeService) parseGeocodingInput(event domain.LambdaEvent) (postalCode, municipality string, err error) {
	var body domain.RequestBody
	if err := json.Unmarshal([]byte(event.Body), &body); err != nil {
		serviceLogger.Error(errorMessageFailedToParseRequestBody, err, nil)
		return "", "", domain.NewValidationError(errorMessageInvalidJSONInRequestBody, "body")
	}

	if body.PostalCode != nil {
		postalCode = *body.PostalCode
	}
	if body.Municipality != nil {
		municipality = *body.Municipality
	}

	return postalCode, municipality, nil
}

// tryGeocodeByPostalCode attempts to geocode by postal code.
// Returns error if postal code format is invalid or not found.
func (s *PostalCodeService) tryGeocodeByPostalCode(postalCode, municipality string) (domain.GeocodingResult, error) {
	// Validate postal code format
	if !postalCodeRegex.MatchString(postalCode) {
		serviceLogger.Warn("Invalid postal code format", map[string]interface{}{
			"postalCode": postalCode,
		})
		return domain.GeocodingResult{}, domain.NewValidationError(
			"Invalid postal code format: expected 5 digits",
			"postalCode",
		)
	}

	result, err := s.provider.GeocodeByPostalCode(postalCode)
	if err != nil {
		if _, ok := err.(*domain.PostalCodeNotFoundError); ok && municipality != "" {
			serviceLogger.Debug("Postal code not found, will try municipality", map[string]interface{}{
				"postalCode":   postalCode,
				"municipality": municipality,
			})
			return domain.GeocodingResult{}, err
		}
		return domain.GeocodingResult{}, err
	}

	serviceLogger.Info("Geocoding successful by postal code", map[string]interface{}{
		"postalCode":   postalCode,
		"municipality": result.Municipality,
		"province":     result.Province,
	})

	return result, nil
}

// geocodeByMunicipality geocodes by municipality name.
func (s *PostalCodeService) geocodeByMunicipality(municipality string) (domain.GeocodingResult, error) {
	result, err := s.provider.GeocodeByMunicipality(municipality)
	if err != nil {
		serviceLogger.Warn("Municipality not found", map[string]interface{}{
			"municipality": municipality,
		})
		return domain.GeocodingResult{}, err
	}

	serviceLogger.Info("Geocoding successful by municipality", map[string]interface{}{
		"municipality": municipality,
		"postalCode":   result.PostalCode,
		"province":     result.Province,
	})

	return result, nil
}

// ReverseGeocode finds the nearest postal code from GPS coordinates.
//
// Uses Haversine distance formula to calculate the great-circle distance
// between the provided coordinates and all postal codes in the database.
// Returns the nearest postal code with its distance in kilometers.
//
// Performance: O(n) brute force search through all postal codes (~10-20ms).
//
// Parameters:
//   - event: LambdaEvent containing lat and lon in the body
//
// Returns:
//   - ReverseGeocodingResult with nearest postal code and distance
//   - error if coordinates are invalid or parsing fails
func (s *PostalCodeService) ReverseGeocode(event domain.LambdaEvent) (domain.ReverseGeocodingResult, error) {
	var body domain.RequestBody
	if err := json.Unmarshal([]byte(event.Body), &body); err != nil {
		serviceLogger.Error(errorMessageFailedToParseRequestBody, err, nil)
		return domain.ReverseGeocodingResult{}, domain.NewValidationError(errorMessageInvalidJSONInRequestBody, "body")
	}

	if body.Lat == nil || body.Lon == nil {
		return domain.ReverseGeocodingResult{}, domain.NewValidationError("lat and lon are required", "coordinates")
	}

	lat := *body.Lat
	lon := *body.Lon

	if err := s.validateCoordinates(lat, lon); err != nil {
		return domain.ReverseGeocodingResult{}, err
	}

	serviceLogger.Debug("Reverse geocoding", map[string]interface{}{
		"lat": lat,
		"lon": lon,
	})

	result, distance, err := s.provider.ReverseGeocode(lat, lon)
	if err != nil {
		serviceLogger.Error("Reverse geocoding failed", err, map[string]interface{}{
			"lat": lat,
			"lon": lon,
		})
		return domain.ReverseGeocodingResult{}, err
	}

	serviceLogger.Info("Reverse geocoding successful", map[string]interface{}{
		"lat":          lat,
		"lon":          lon,
		"postalCode":   result.PostalCode,
		"municipality": result.Municipality,
		"distance":     distance,
	})

	return domain.ReverseGeocodingResult{
		Success:    result.Success,
		City:       result.Municipality,
		PostalCode: result.PostalCode,
		Province:   result.Province,
		Country:    "Portugal",
		Coords:     result.Coords,
		Distance:   distance,
	}, nil
}

// ValidatePostal validates if a postal code exists in the database.
//
// Performs an O(1) lookup in the postal codes map.
// Very fast validation (<1ms latency).
//
// Parameters:
//   - event: LambdaEvent containing postalCode in the body
//
// Returns:
//   - ValidationResult with valid flag and the postal code value
//   - error if input parsing fails
func (s *PostalCodeService) ValidatePostal(event domain.LambdaEvent) (domain.ValidationResult, error) {
	postalCode, err := s.parseValidationInput(event, "postalCode")
	if err != nil {
		return domain.ValidationResult{}, err
	}

	serviceLogger.Debug("Validating postal code", map[string]interface{}{
		"postalCode": postalCode,
	})

	isValid := s.provider.ValidatePostalCode(postalCode)

	serviceLogger.Info("Postal code validation result", map[string]interface{}{
		"postalCode": postalCode,
		"valid":      isValid,
	})

	return domain.ValidationResult{
		Valid: isValid,
		Value: postalCode,
	}, nil
}

// ValidateMunicipality validates if a municipality exists in the database.
//
// Performs an O(1) lookup in the municipality set using lowercase normalized name.
// Very fast validation (<1ms latency).
//
// Parameters:
//   - event: LambdaEvent containing municipality in the body
//
// Returns:
//   - ValidationResult with valid flag and the municipality value
//   - error if input parsing fails
func (s *PostalCodeService) ValidateMunicipality(event domain.LambdaEvent) (domain.ValidationResult, error) {
	municipality, err := s.parseValidationInput(event, "municipality")
	if err != nil {
		return domain.ValidationResult{}, err
	}

	serviceLogger.Debug("Validating municipality", map[string]interface{}{
		"municipality": municipality,
	})

	isValid := s.provider.ValidateMunicipality(municipality)

	serviceLogger.Info("Municipality validation result", map[string]interface{}{
		"municipality": municipality,
		"valid":        isValid,
	})

	return domain.ValidationResult{
		Valid: isValid,
		Value: municipality,
	}, nil
}

// AutocompletePostal returns postal codes matching the given prefix.
//
// Uses binary search on a sorted array of postal codes for efficient prefix matching.
// Performance: O(log n) to find first match + O(k) where k is result count.
//
// Parameters:
//   - event: LambdaEvent containing prefix and optional limit in the body
//
// Returns:
//   - Slice of AutocompleteResult with matching postal codes (up to limit)
//   - error if input parsing fails
//
// Limits:
//   - Default limit: 10 results
//   - Maximum limit: 50 results (enforced automatically)
func (s *PostalCodeService) AutocompletePostal(event domain.LambdaEvent) ([]domain.AutocompleteResult, error) {
	prefix, limit, err := s.parseAutocompleteInput(event, true)
	if err != nil {
		return nil, err
	}

	serviceLogger.Debug("Autocomplete postal code", map[string]interface{}{
		"prefix": prefix,
		"limit":  limit,
	})

	results := s.provider.AutocompletePostalCode(prefix, limit)

	serviceLogger.Info("Autocomplete postal code results", map[string]interface{}{
		"prefix":       prefix,
		"limit":        limit,
		"resultsCount": len(results),
	})

	return results, nil
}

// AutocompleteMunicipality returns municipalities matching the given query (fuzzy search).
//
// Performs a case-insensitive search through the municipality index, prioritizing
// results that start with the query over those that contain it. Results are
// sorted alphabetically with starts-with matches first.
//
// Performance: O(n) search through municipality index (~5-10ms).
//
// Parameters:
//   - event: LambdaEvent containing query and optional limit in the body
//
// Returns:
//   - Slice of AutocompleteResult with matching municipalities (up to limit)
//   - error if input parsing fails
//
// Limits:
//   - Default limit: 10 results
//   - Maximum limit: 50 results (enforced automatically)
func (s *PostalCodeService) AutocompleteMunicipality(event domain.LambdaEvent) ([]domain.AutocompleteResult, error) {
	query, limit, err := s.parseAutocompleteInput(event, false)
	if err != nil {
		return nil, err
	}

	serviceLogger.Debug("Autocomplete municipality", map[string]interface{}{
		"query": query,
		"limit": limit,
	})

	results := s.provider.AutocompleteMunicipality(query, limit)

	serviceLogger.Info("Autocomplete municipality results", map[string]interface{}{
		"query":        query,
		"limit":        limit,
		"resultsCount": len(results),
	})

	return results, nil
}

// parseValidationInput parses and validates validation input from Lambda event.
func (s *PostalCodeService) parseValidationInput(event domain.LambdaEvent, field string) (string, error) {
	var body domain.RequestBody
	if err := json.Unmarshal([]byte(event.Body), &body); err != nil {
		serviceLogger.Error(errorMessageFailedToParseRequestBody, err, nil)
		return "", domain.NewValidationError(errorMessageInvalidJSONInRequestBody, "body")
	}

	var value string
	if field == "postalCode" && body.PostalCode != nil {
		value = *body.PostalCode
	} else if field == "municipality" && body.Municipality != nil {
		value = *body.Municipality
	}

	if value == "" {
		return "", domain.NewValidationError(
			field+" is required and must be a non-empty string",
			field,
		)
	}

	return value, nil
}

// parseAutocompleteInput parses and validates autocomplete input from Lambda event.
// isPostalCode indicates whether this is for postal code (prefix) or municipality (query).
func (s *PostalCodeService) parseAutocompleteInput(event domain.LambdaEvent, isPostalCode bool) (string, int, error) {
	var body domain.RequestBody
	if err := json.Unmarshal([]byte(event.Body), &body); err != nil {
		serviceLogger.Error(errorMessageFailedToParseRequestBody, err, nil)
		return "", 0, domain.NewValidationError(errorMessageInvalidJSONInRequestBody, "body")
	}

	var value string
	if isPostalCode {
		if body.Prefix == nil {
			return "", 0, domain.NewValidationError("prefix is required", "prefix")
		}
		value = *body.Prefix
	} else {
		if body.Query == nil {
			return "", 0, domain.NewValidationError("query is required", "query")
		}
		value = *body.Query
	}

	limit := domain.DefaultAutocompleteLimit
	if body.Limit != nil {
		limit = *body.Limit
		if limit < 1 {
			return "", 0, domain.NewValidationError("limit must be a positive number", "limit")
		}
		if limit > domain.MaxAutocompleteLimit {
			limit = domain.MaxAutocompleteLimit
		}
	}

	return value, limit, nil
}
