// Package provider contains data providers for postal code information.
package provider

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"

	"github.com/pricofy/geocode-pt/internal/domain"
	"github.com/pricofy/geocode-pt/internal/shared/logger"
)

//go:embed postal-codes-pt.json
var postalCodesJSON []byte

// providerLogger is the logger instance for the provider package
var providerLogger = logger.NewLogger("PostalCodeProvider")

// PostalCodeProvider provides access to Portuguese postal code data.
// Implements lazy loading with sync.Once to ensure single initialization.
// Builds optimized indices for O(1) lookups and fast autocomplete.
type PostalCodeProvider struct {
	codes             map[string]domain.PostalData
	municipalityIndex map[string][]postalEntry
	municipalitySet   map[string]bool
	sortedPostalCodes []string
	initOnce          sync.Once
}

// postalEntry represents a postal code entry in the municipality index.
type postalEntry struct {
	PostalCode string
	Data       domain.PostalData
}

// NewPostalCodeProvider creates a new PostalCodeProvider instance.
// The database is loaded lazily on first access.
func NewPostalCodeProvider() *PostalCodeProvider {
	return &PostalCodeProvider{}
}

// getCodes loads the postal codes database and builds indices if not already loaded.
// Uses sync.Once to ensure thread-safe single initialization.
func (p *PostalCodeProvider) getCodes() map[string]domain.PostalData {
	p.initOnce.Do(func() {
		var codes map[string]domain.PostalData
		if err := json.Unmarshal(postalCodesJSON, &codes); err != nil {
			providerLogger.Error("Failed to parse postal codes JSON", err, nil)
			panic(fmt.Sprintf("Failed to load postal codes database: %v", err))
		}

		p.codes = codes
		p.buildIndexes(codes)

		providerLogger.Info("Loaded postal codes database with indexes", map[string]interface{}{
			"totalPostalCodes":    len(codes),
			"uniqueMunicipalitys": len(p.municipalitySet),
		})
	})

	return p.codes
}

// buildIndexes builds optimized indices for fast lookups:
// - municipalityIndex: map of lowercase municipality → []postalEntry
// - municipalitySet: set of lowercase municipality names for O(1) validation
// - sortedPostalCodes: sorted array for binary search autocomplete
func (p *PostalCodeProvider) buildIndexes(codes map[string]domain.PostalData) {
	p.municipalityIndex = make(map[string][]postalEntry)
	p.municipalitySet = make(map[string]bool)
	p.sortedPostalCodes = make([]string, 0, len(codes))

	for postalCode, data := range codes {
		municipalityKey := strings.ToLower(strings.TrimSpace(data.Municipality))
		p.municipalitySet[municipalityKey] = true

		if p.municipalityIndex[municipalityKey] == nil {
			p.municipalityIndex[municipalityKey] = make([]postalEntry, 0)
		}
		p.municipalityIndex[municipalityKey] = append(p.municipalityIndex[municipalityKey], postalEntry{
			PostalCode: postalCode,
			Data:       data,
		})

		p.sortedPostalCodes = append(p.sortedPostalCodes, postalCode)
	}

	sort.Strings(p.sortedPostalCodes)

	providerLogger.Debug("Built indexes", map[string]interface{}{
		"municipalityIndexSize":   len(p.municipalityIndex),
		"municipalitySetSize":     len(p.municipalitySet),
		"sortedPostalCodesLength": len(p.sortedPostalCodes),
	})
}

// calculateDistance calculates the great-circle distance between two points on Earth
// using the Haversine formula.
//
// The Haversine formula determines the great-circle distance between two points
// on a sphere given their longitudes and latitudes. This is the shortest distance
// over the Earth's surface, ignoring terrain.
//
// Parameters:
//   - lat1, lon1: Coordinates of the first point (in degrees)
//   - lat2, lon2: Coordinates of the second point (in degrees)
//
// Returns:
//   - Distance in kilometers between the two points
func (p *PostalCodeProvider) calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = domain.EarthRadiusKm // Earth radius in kilometers

	dLat := (lat2 - lat1) * (math.Pi / 180)
	dLon := (lon2 - lon1) * (math.Pi / 180)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*(math.Pi/180))*
			math.Cos(lat2*(math.Pi/180))*
			math.Sin(dLon/2)*
			math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

// GeocodeByPostalCode looks up a postal code in the database and returns coordinates.
//
// Performs an O(1) lookup in the postal codes map using the postal code as key.
// Very fast operation (<1ms latency).
//
// Parameters:
//   - postalCode: Portuguese postal code (7 digits, e.g., "1000205")
//
// Returns:
//   - GeocodingResult with coordinates, municipality, and province
//   - PostalCodeNotFoundError if postal code doesn't exist
func (p *PostalCodeProvider) GeocodeByPostalCode(postalCode string) (domain.GeocodingResult, error) {
	providerLogger.Debug("Geocoding by postal code", map[string]interface{}{
		"postalCode": postalCode,
	})

	codes := p.getCodes()
	data, exists := codes[postalCode]
	if !exists {
		providerLogger.Warn("Postal code not found", map[string]interface{}{
			"postalCode": postalCode,
		})
		return domain.GeocodingResult{}, domain.NewPostalCodeNotFoundError(postalCode, "")
	}

	providerLogger.Info("Geocoding successful", map[string]interface{}{
		"postalCode":   postalCode,
		"municipality": data.Municipality,
		"province":     data.Province,
		"lat":          data.Lat,
		"lon":          data.Lon,
		"source":       "postal_code",
	})

	return domain.GeocodingResult{
		Success:      true,
		Coords:       domain.Coordinates{Lat: data.Lat, Lon: data.Lon},
		Municipality: data.Municipality,
		Province:     data.Province,
		PostalCode:   postalCode,
		Source:       "postal_code",
	}, nil
}

// GeocodeByMunicipality finds a postal code by municipality name.
//
// Performs a case-insensitive linear search through all postal codes to find
// the first match for the given municipality name. Returns the first postal
// code found for that municipality.
//
// Performance: O(n) search through all postal codes (~5ms).
//
// Parameters:
//   - municipality: Municipality name (case-insensitive, e.g., "Madrid" or "madrid")
//
// Returns:
//   - GeocodingResult with coordinates, postal code, and province
//   - PostalCodeNotFoundError if municipality doesn't exist
func (p *PostalCodeProvider) GeocodeByMunicipality(municipality string) (domain.GeocodingResult, error) {
	providerLogger.Debug("Geocoding by municipality", map[string]interface{}{
		"municipality": municipality,
	})

	codes := p.getCodes()
	municipalityLower := strings.ToLower(strings.TrimSpace(municipality))

	// O(n) search through all postal codes
	for postalCode, data := range codes {
		if strings.ToLower(data.Municipality) == municipalityLower {
			providerLogger.Info("Geocoding successful", map[string]interface{}{
				"municipality": municipality,
				"postalCode":   postalCode,
				"province":     data.Province,
				"lat":          data.Lat,
				"lon":          data.Lon,
				"source":       "municipality",
			})

			return domain.GeocodingResult{
				Success:      true,
				Coords:       domain.Coordinates{Lat: data.Lat, Lon: data.Lon},
				Municipality: data.Municipality,
				Province:     data.Province,
				PostalCode:   postalCode,
				Source:       "municipality",
			}, nil
		}
	}

	providerLogger.Warn("Municipality not found", map[string]interface{}{
		"municipality": municipality,
	})
	return domain.GeocodingResult{}, domain.NewPostalCodeNotFoundError("", municipality)
}

// ReverseGeocode finds the nearest postal code from GPS coordinates using Haversine distance.
//
// Calculates the great-circle distance between the provided coordinates and all
// postal codes in the database using the Haversine formula. Returns the nearest
// postal code with its distance in kilometers (rounded to 3 decimal places).
//
// Performance: O(n) brute force search through all 11,150 postal codes (~10-20ms).
//
// Parameters:
//   - lat: Latitude (-90 to 90)
//   - lon: Longitude (-180 to 180)
//
// Returns:
//   - GeocodingResult with nearest postal code data
//   - distance: Distance in kilometers to the nearest postal code
//   - error if no postal code found (should never happen with valid data)
func (p *PostalCodeProvider) ReverseGeocode(lat, lon float64) (domain.GeocodingResult, float64, error) {
	providerLogger.Debug("Reverse geocoding", map[string]interface{}{
		"lat": lat,
		"lon": lon,
	})

	codes := p.getCodes()

	// Find nearest postal code (brute force)
	var nearest *struct {
		postalCode string
		data       domain.PostalData
		distance   float64
	}
	minDistance := math.MaxFloat64

	for postalCode, data := range codes {
		dist := p.calculateDistance(lat, lon, data.Lat, data.Lon)

		if dist < minDistance {
			minDistance = dist
			nearest = &struct {
				postalCode string
				data       domain.PostalData
				distance   float64
			}{
				postalCode: postalCode,
				data:       data,
				distance:   dist,
			}
		}
	}

	if nearest == nil {
		providerLogger.Error("No postal code found (should never happen)", nil, nil)
		return domain.GeocodingResult{}, 0, domain.NewPostalCodeNotFoundError("", "")
	}

	// Round distance to 3 decimal places
	distance := math.Round(minDistance*1000) / 1000

	providerLogger.Info("Reverse geocoding successful", map[string]interface{}{
		"lat":          lat,
		"lon":          lon,
		"postalCode":   nearest.postalCode,
		"municipality": nearest.data.Municipality,
		"province":     nearest.data.Province,
		"distance":     distance,
	})

	return domain.GeocodingResult{
		Success:      true,
		Coords:       domain.Coordinates{Lat: nearest.data.Lat, Lon: nearest.data.Lon},
		Municipality: nearest.data.Municipality,
		Province:     nearest.data.Province,
		PostalCode:   nearest.postalCode,
		Source:       "reverse_geocode",
	}, distance, nil
}

// ValidatePostalCode checks if a postal code exists in the database.
//
// Performs an O(1) lookup in the postal codes map.
// Very fast validation (<1ms latency).
//
// Parameters:
//   - postalCode: Portuguese postal code to validate (7 digits)
//
// Returns:
//   - true if postal code exists in the database
//   - false if postal code doesn't exist
func (p *PostalCodeProvider) ValidatePostalCode(postalCode string) bool {
	providerLogger.Debug("Validating postal code", map[string]interface{}{
		"postalCode": postalCode,
	})

	codes := p.getCodes()
	exists := false
	if _, ok := codes[postalCode]; ok {
		exists = true
	}

	providerLogger.Debug("Postal code validation result", map[string]interface{}{
		"postalCode": postalCode,
		"exists":     exists,
	})

	return exists
}

// ValidateMunicipality checks if a municipality exists in the database.
//
// Performs an O(1) lookup in the municipality set using lowercase normalized name.
// Very fast validation (<1ms latency).
//
// Parameters:
//   - municipality: Municipality name to validate (case-insensitive)
//
// Returns:
//   - true if municipality exists in the database
//   - false if municipality doesn't exist
func (p *PostalCodeProvider) ValidateMunicipality(municipality string) bool {
	providerLogger.Debug("Validating municipality", map[string]interface{}{
		"municipality": municipality,
	})

	p.getCodes() // Ensure indexes are built
	municipalityLower := strings.ToLower(strings.TrimSpace(municipality))
	exists := p.municipalitySet[municipalityLower]

	providerLogger.Debug("Municipality validation result", map[string]interface{}{
		"municipality": municipality,
		"exists":       exists,
	})

	return exists
}

// AutocompletePostalCode returns postal codes matching the given prefix.
//
// Uses binary search on a pre-sorted array of postal codes to efficiently find
// all codes that start with the given prefix. Results are returned in sorted order.
//
// Performance: O(log n) to find first match + O(k) where k is the number of results.
//
// Parameters:
//   - prefix: Postal code prefix to search for (e.g., "280" for Madrid codes)
//   - limit: Maximum number of results to return
//
// Returns:
//   - Slice of AutocompleteResult with matching postal codes (up to limit)
//   - Empty slice if no matches found
func (p *PostalCodeProvider) AutocompletePostalCode(prefix string, limit int) []domain.AutocompleteResult {
	providerLogger.Debug("Autocomplete postal code", map[string]interface{}{
		"prefix": prefix,
		"limit":  limit,
	})

	codes := p.getCodes()
	results := make([]domain.AutocompleteResult, 0)

	// Binary search for first matching prefix
	startIndex := sort.Search(len(p.sortedPostalCodes), func(i int) bool {
		return p.sortedPostalCodes[i] >= prefix
	})

	if startIndex >= len(p.sortedPostalCodes) {
		providerLogger.Debug("No postal codes found for prefix", map[string]interface{}{
			"prefix": prefix,
		})
		return results
	}

	// Collect matching postal codes
	for i := startIndex; i < len(p.sortedPostalCodes) && len(results) < limit; i++ {
		postalCode := p.sortedPostalCodes[i]
		if !strings.HasPrefix(postalCode, prefix) {
			break
		}

		data := codes[postalCode]
		results = append(results, domain.AutocompleteResult{
			PostalCode:   postalCode,
			Municipality: data.Municipality,
			Province:     data.Province,
		})
	}

	providerLogger.Info("Autocomplete postal code results", map[string]interface{}{
		"prefix":       prefix,
		"limit":        limit,
		"resultsCount": len(results),
	})

	return results
}

// AutocompleteMunicipality returns municipalities matching the given query (fuzzy search).
//
// Performs a case-insensitive search through the municipality index, matching
// municipalities that either start with or contain the query string. Results
// are sorted with starts-with matches first, then alphabetically.
//
// Performance: O(n) search through municipality index (~5-10ms).
//
// Parameters:
//   - query: Municipality name query (case-insensitive, e.g., "mad" matches "Madrid")
//   - limit: Maximum number of results to return
//
// Returns:
//   - Slice of AutocompleteResult with matching municipalities (up to limit)
//   - Empty slice if no matches found
func (p *PostalCodeProvider) AutocompleteMunicipality(query string, limit int) []domain.AutocompleteResult {
	providerLogger.Debug("Autocomplete municipality", map[string]interface{}{
		"query": query,
		"limit": limit,
	})

	p.getCodes() // Ensure indexes are built
	queryLower := strings.ToLower(strings.TrimSpace(query))
	results := make([]domain.AutocompleteResult, 0)
	seenMunicipalities := make(map[string]bool)

	// Collect matches (prioritize starts-with)
	type match struct {
		result     domain.AutocompleteResult
		startsWith bool
	}
	matches := make([]match, 0)

	for municipalityKey, entries := range p.municipalityIndex {
		if len(matches) >= limit*2 { // Collect more for sorting
			break
		}

		startsWith := strings.HasPrefix(municipalityKey, queryLower)
		contains := strings.Contains(municipalityKey, queryLower)

		if startsWith || contains {
			if !seenMunicipalities[municipalityKey] {
				seenMunicipalities[municipalityKey] = true
				entry := entries[0] // Use first postal code for this municipality
				matches = append(matches, match{
					result: domain.AutocompleteResult{
						PostalCode:   entry.PostalCode,
						Municipality: entry.Data.Municipality,
						Province:     entry.Data.Province,
					},
					startsWith: startsWith,
				})
			}
		}
	}

	// Sort: starts-with first, then alphabetically
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].startsWith && !matches[j].startsWith {
			return true
		}
		if !matches[i].startsWith && matches[j].startsWith {
			return false
		}
		return strings.ToLower(matches[i].result.Municipality) < strings.ToLower(matches[j].result.Municipality)
	})

	// Take up to limit results
	for i := 0; i < len(matches) && i < limit; i++ {
		results = append(results, matches[i].result)
	}

	providerLogger.Info("Autocomplete municipality results", map[string]interface{}{
		"query":        query,
		"limit":        limit,
		"resultsCount": len(results),
	})

	return results
}

// GeocodeByMunicipalitiesBatch geocodes multiple municipalities in a single batch operation.
func (p *PostalCodeProvider) GeocodeByMunicipalitiesBatch(municipalities []string) map[string]*domain.BatchGeocodingResult {
	providerLogger.Debug("Batch geocoding municipalities", map[string]interface{}{"count": len(municipalities)})
	p.getCodes()
	results := make(map[string]*domain.BatchGeocodingResult, len(municipalities))

	for _, municipality := range municipalities {
		municipalityLower := strings.ToLower(strings.TrimSpace(municipality))
		if entries, ok := p.municipalityIndex[municipalityLower]; ok && len(entries) > 0 {
			entry := entries[0]
			results[municipality] = &domain.BatchGeocodingResult{
				Lat: entry.Data.Lat, Lon: entry.Data.Lon, Found: true, PostalCode: entry.PostalCode,
			}
		} else {
			results[municipality] = nil
		}
	}

	foundCount := 0
	for _, result := range results {
		if result != nil && result.Found {
			foundCount++
		}
	}
	providerLogger.Info("Batch geocoding completed", map[string]interface{}{"requested": len(municipalities), "found": foundCount})
	return results
}
