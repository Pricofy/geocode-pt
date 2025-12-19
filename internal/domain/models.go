package domain

// PostalData represents a Spanish postal code entry in the database.
// Contains geographic coordinates, municipality, and province information.
type PostalData struct {
	Lat          float64 `json:"lat"`
	Lon          float64 `json:"lon"`
	Municipality string  `json:"municipality"`
	Province     string  `json:"province"`
}

// Coordinates represents geographic coordinates (latitude and longitude).
type Coordinates struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// GeocodingResult represents the result of a geocoding operation
// (postal code or municipality → coordinates).
type GeocodingResult struct {
	Success      bool        `json:"success"`
	Coords       Coordinates `json:"coords"`
	Municipality string      `json:"municipality"`
	Province     string      `json:"province"`
	PostalCode   string      `json:"postalCode"`
	Source       string      `json:"source"` // "postal_code", "municipality", or "reverse_geocode"
}

// ReverseGeocodingResult represents the result of a reverse geocoding operation
// (coordinates → nearest postal code).
type ReverseGeocodingResult struct {
	Success    bool        `json:"success"`
	City       string      `json:"city"` // Municipality name
	PostalCode string      `json:"postalCode"`
	Province   string      `json:"province"`
	Country    string      `json:"country"` // Always "España" for this service
	Coords     Coordinates `json:"coords"`
	Distance   float64     `json:"distance"` // Distance in kilometers
}

// AutocompleteResult represents a single result in an autocomplete operation.
type AutocompleteResult struct {
	PostalCode   string `json:"postalCode"`
	Municipality string `json:"municipality"`
	Province     string `json:"province"`
}

// ValidationResult represents the result of a validation operation.
type ValidationResult struct {
	Valid bool   `json:"valid"`
	Value string `json:"value"`
}

// LambdaEvent represents the Lambda function invocation event.
// The Body field contains the JSON-encoded request payload.
type LambdaEvent struct {
	Body string `json:"body"`
}

// LambdaResponse represents the Lambda function response.
// StatusCode follows HTTP status codes, Body is JSON-encoded.
type LambdaResponse struct {
	StatusCode int    `json:"statusCode"`
	Body       string `json:"body"`
}

// RequestBody represents the parsed request body from LambdaEvent.
// Used for routing operations based on the "operation" field.
type RequestBody struct {
	Operation       string   `json:"operation"`
	PostalCode      *string  `json:"postalCode,omitempty"`
	Municipality    *string  `json:"municipality,omitempty"`
	Lat             *float64 `json:"lat,omitempty"`
	Lon             *float64 `json:"lon,omitempty"`
	Prefix          *string  `json:"prefix,omitempty"`
	Query           *string  `json:"query,omitempty"`
	Limit           *int     `json:"limit,omitempty"`
	Municipalities  []string `json:"municipalities,omitempty"`
}

// BatchGeocodingResult represents the result of geocoding a single municipality in a batch operation.
type BatchGeocodingResult struct {
	Lat        float64 `json:"lat"`
	Lon        float64 `json:"lon"`
	Found      bool    `json:"found"`
	PostalCode string  `json:"postalCode"`
}

// BatchGeocodingResponse represents the response for a batch geocoding operation.
type BatchGeocodingResponse struct {
	Success bool                             `json:"success"`
	Results map[string]*BatchGeocodingResult `json:"results"`
	Count   int                              `json:"count"`
	Errors  []string                         `json:"errors"`
}

