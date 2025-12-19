// Package main is the entry point for the Pricofy Geocode ES Lambda function.
// It initializes services and starts the Lambda handler using dependency injection.
//
// The Lambda function supports the following operations:
//   - geocode-by-postal: Convert postal code to coordinates
//   - reverse-geocode: Find nearest postal code from coordinates
//   - validate-postal: Validate if postal code exists
//   - validate-municipality: Validate if municipality exists
//   - autocomplete-postal: Autocomplete postal codes by prefix
//   - autocomplete-municipality: Autocomplete municipalities by query
package main

import (
	"context"
	"encoding/json"
	"fmt"

	awslambda "github.com/aws/aws-lambda-go/lambda"
	"github.com/pricofy/geocode-pt/internal/application"
	"github.com/pricofy/geocode-pt/internal/domain"
	"github.com/pricofy/geocode-pt/internal/shared/logger"
)

// App encapsulates all application dependencies.
// This struct enables dependency injection and makes the application fully testable
// by allowing mock implementations to be provided in tests.
type App struct {
	service *application.PostalCodeService
	logger  *logger.Logger
}

// NewApp creates a new App instance with all dependencies injected.
// This constructor follows the dependency injection principle, making dependencies
// explicit and allowing easy substitution of mocks in tests.
//
// Returns:
//   - *App: Initialized application with all dependencies wired
//   - error: Initialization error if any component fails
//
//nolint:unparam // error return is for future extensibility
func NewApp() (*App, error) {
	// Initialize PostalCodeService
	service := application.NewPostalCodeService()

	// Initialize Logger
	appLogger := logger.NewLogger("GeocodeES")

	return &App{
		service: service,
		logger:  appLogger,
	}, nil
}

// Global app instance - initialized once for Lambda cold start optimization.
var globalApp *App

// init initializes the global app instance on Lambda cold start.
func init() {
	app, err := NewApp()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize application: %v", err))
	}

	globalApp = app
}

// main starts the Lambda runtime.
func main() {
	awslambda.Start(Handler)
}

// Handler is the main Lambda handler function.
// It routes incoming events to the appropriate operation handler based on the "operation" field.
// Uses the global app instance initialized during cold start.
func Handler(ctx context.Context, event domain.LambdaEvent) (domain.LambdaResponse, error) {
	return globalApp.HandleRequest(ctx, event)
}

// HandleRequest processes a Lambda invocation request.
// This method is separated from Handler to enable testing without AWS Lambda runtime.
//
// Parameters:
//   - ctx: Request context with timeout and cancellation
//   - event: Lambda event payload with "operation" field in body
//
// Returns:
//   - domain.LambdaResponse: Response with statusCode and body
//   - error: Error if critical failure (Lambda will retry)
func (a *App) HandleRequest(ctx context.Context, event domain.LambdaEvent) (domain.LambdaResponse, error) {
	// Parse request body
	var requestBody domain.RequestBody
	if err := json.Unmarshal([]byte(event.Body), &requestBody); err != nil {
		a.logger.Error("Failed to parse request body", err, map[string]interface{}{
			"body": event.Body,
		})
		return domain.LambdaResponse{
			StatusCode: 400,
			Body:       `{"success":false,"error":"Invalid JSON in request body"}`,
		}, nil
	}

	operation := requestBody.Operation
	a.logger.Info(fmt.Sprintf("Processing operation: %s", operation), map[string]interface{}{
		"operation": operation,
	})

	// Route to appropriate handler based on operation
	switch operation {
	case "geocode-by-postal":
		return a.handleGeocodeByPostal(ctx, event)
	case "reverse-geocode":
		return a.handleReverseGeocode(ctx, event)
	case "validate-postal":
		return a.handleValidatePostal(ctx, event)
	case "validate-municipality":
		return a.handleValidateMunicipality(ctx, event)
	case "autocomplete-postal":
		return a.handleAutocompletePostal(ctx, event)
	case "autocomplete-municipality":
		return a.handleAutocompleteMunicipality(ctx, event)
	case "geocode-municipalities-batch":
		return a.handleGeocodeMunicipalitiesBatch(ctx, event)
	default:
		a.logger.Warn(fmt.Sprintf("Unknown operation requested: %s", operation), map[string]interface{}{
			"operation": operation,
		})
		return domain.LambdaResponse{
			StatusCode: 400,
			Body:       fmt.Sprintf(`{"success":false,"error":"Unknown operation: %s"}`, operation),
		}, nil
	}
}

// handleGeocodeByPostal handles geocode-by-postal operation.
func (a *App) handleGeocodeByPostal(_ context.Context, event domain.LambdaEvent) (domain.LambdaResponse, error) {
	result, err := a.service.GeocodeByPostal(event)
	if err != nil {
		if _, ok := err.(*domain.PostalCodeNotFoundError); ok {
			body, _ := json.Marshal(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
			return domain.LambdaResponse{
				StatusCode: 404,
				Body:       string(body),
			}, nil
		}
		if _, ok := err.(*domain.ValidationError); ok {
			body, _ := json.Marshal(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
			return domain.LambdaResponse{
				StatusCode: 400,
				Body:       string(body),
			}, nil
		}
		body, _ := json.Marshal(map[string]interface{}{
			"success": false,
			"error":   "Internal server error",
		})
		return domain.LambdaResponse{
			StatusCode: 500,
			Body:       string(body),
		}, nil
	}

	body, _ := json.Marshal(result)
	return domain.LambdaResponse{
		StatusCode: 200,
		Body:       string(body),
	}, nil
}

// handleReverseGeocode handles reverse-geocode operation.
func (a *App) handleReverseGeocode(_ context.Context, event domain.LambdaEvent) (domain.LambdaResponse, error) {
	result, err := a.service.ReverseGeocode(event)
	if err != nil {
		if _, ok := err.(*domain.InvalidCoordinatesError); ok {
			body, _ := json.Marshal(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
			return domain.LambdaResponse{
				StatusCode: 400,
				Body:       string(body),
			}, nil
		}
		body, _ := json.Marshal(map[string]interface{}{
			"success": false,
			"error":   "Internal server error",
		})
		return domain.LambdaResponse{
			StatusCode: 500,
			Body:       string(body),
		}, nil
	}

	body, _ := json.Marshal(result)
	return domain.LambdaResponse{
		StatusCode: 200,
		Body:       string(body),
	}, nil
}

// handleValidatePostal handles validate-postal operation.
func (a *App) handleValidatePostal(_ context.Context, event domain.LambdaEvent) (domain.LambdaResponse, error) {
	result, err := a.service.ValidatePostal(event)
	if err != nil {
		body, _ := json.Marshal(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return domain.LambdaResponse{
			StatusCode: 400,
			Body:       string(body),
		}, nil
	}

	body, _ := json.Marshal(result)
	return domain.LambdaResponse{
		StatusCode: 200,
		Body:       string(body),
	}, nil
}

// handleValidateMunicipality handles validate-municipality operation (validate municipality).
func (a *App) handleValidateMunicipality(_ context.Context, event domain.LambdaEvent) (domain.LambdaResponse, error) {
	result, err := a.service.ValidateMunicipality(event)
	if err != nil {
		body, _ := json.Marshal(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return domain.LambdaResponse{
			StatusCode: 400,
			Body:       string(body),
		}, nil
	}

	body, _ := json.Marshal(result)
	return domain.LambdaResponse{
		StatusCode: 200,
		Body:       string(body),
	}, nil
}

// handleAutocompletePostal handles autocomplete-postal operation.
func (a *App) handleAutocompletePostal(_ context.Context, event domain.LambdaEvent) (domain.LambdaResponse, error) {
	results, err := a.service.AutocompletePostal(event)
	if err != nil {
		body, _ := json.Marshal(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return domain.LambdaResponse{
			StatusCode: 400,
			Body:       string(body),
		}, nil
	}

	body, _ := json.Marshal(map[string]interface{}{
		"success": true,
		"results": results,
	})
	return domain.LambdaResponse{
		StatusCode: 200,
		Body:       string(body),
	}, nil
}

// handleAutocompleteMunicipality handles autocomplete-municipality operation (autocomplete municipality).
func (a *App) handleAutocompleteMunicipality(_ context.Context, event domain.LambdaEvent) (domain.LambdaResponse, error) {
	results, err := a.service.AutocompleteMunicipality(event)
	if err != nil {
		body, _ := json.Marshal(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return domain.LambdaResponse{
			StatusCode: 400,
			Body:       string(body),
		}, nil
	}

	body, _ := json.Marshal(map[string]interface{}{
		"success": true,
		"results": results,
	})
	return domain.LambdaResponse{
		StatusCode: 200,
		Body:       string(body),
	}, nil
}

// handleGeocodeMunicipalitiesBatch handles geocode-municipalities-batch operation.
func (a *App) handleGeocodeMunicipalitiesBatch(_ context.Context, event domain.LambdaEvent) (domain.LambdaResponse, error) {
	result, err := a.service.GeocodeMunicipalitiesBatch(event)
	if err != nil {
		if _, ok := err.(*domain.ValidationError); ok {
			body, _ := json.Marshal(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
			return domain.LambdaResponse{
				StatusCode: 400,
				Body:       string(body),
			}, nil
		}
		body, _ := json.Marshal(map[string]interface{}{
			"success": false,
			"error":   "Internal server error",
		})
		return domain.LambdaResponse{
			StatusCode: 500,
			Body:       string(body),
		}, nil
	}

	body, _ := json.Marshal(result)
	return domain.LambdaResponse{
		StatusCode: 200,
		Body:       string(body),
	}, nil
}
