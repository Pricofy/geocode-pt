// Package handler contains AWS Lambda handler implementations.
package handler

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/pricofy/geocode-pt/internal/application"
	"github.com/pricofy/geocode-pt/internal/domain"
	"github.com/pricofy/geocode-pt/internal/shared/logger"
)

// handlerLogger is the logger instance for the handler
var handlerLogger = logger.NewLogger("LambdaHandler")

// Handler is the main Lambda handler that routes requests to appropriate operations.
// Supports all Spanish postal code operations:
// - geocode-by-postal
// - reverse-geocode
// - validate-postal
// - validate-municipality
// - autocomplete-postal
// - autocomplete-municipality
//
// Handles both event formats:
// 1. Direct Lambda invocation: { "operation": "...", "postalCode": "..." }
// 2. API Gateway format: { "body": "{\"operation\":\"...\"}" }
func Handler(_ context.Context, event interface{}) (domain.LambdaResponse, error) {
	handlerLogger.Info("Processing request", nil)

	// Convert event to JSON bytes for parsing
	eventJSON, err := json.Marshal(event)
	if err != nil {
		handlerLogger.Error("Failed to marshal event", err, nil)
		errorBody, _ := json.Marshal(map[string]interface{}{
			"success": false,
			"error":   "Invalid event format",
		})
		return domain.LambdaResponse{
			StatusCode: 400,
			Body:       string(errorBody),
		}, nil
	}

	// Try to parse as LambdaEvent first (API Gateway format)
	var lambdaEvent domain.LambdaEvent
	if err := json.Unmarshal(eventJSON, &lambdaEvent); err == nil && lambdaEvent.Body != "" {
		// Format 1: API Gateway format with body field
		eventJSON = []byte(lambdaEvent.Body)
	}

	// Parse as RequestBody (direct Lambda invocation or parsed body)
	var body domain.RequestBody
	if err := json.Unmarshal(eventJSON, &body); err != nil {
		handlerLogger.Error("Failed to parse request body", err, nil)
		errorBody, _ := json.Marshal(map[string]interface{}{
			"success": false,
			"error":   "Invalid JSON in request body",
		})
		return domain.LambdaResponse{
			StatusCode: 400,
			Body:       string(errorBody),
		}, nil
	}
	
	// Create a LambdaEvent with the parsed body for operations (normalized format)
	bodyJSON, _ := json.Marshal(body)
	eventWithBody := domain.LambdaEvent{
		Body: string(bodyJSON),
	}

	// Validate operation parameter
	if body.Operation == "" {
		errorBody, _ := json.Marshal(map[string]interface{}{
			"success": false,
			"error":   "Operation parameter is required",
		})
		return domain.LambdaResponse{
			StatusCode: 400,
			Body:       string(errorBody),
		}, nil
	}

	// Create service instance
	service := application.NewPostalCodeService()

	// Route to appropriate operation
	switch body.Operation {
	case "geocode-by-postal":
		return application.GeocodeByPostalOperation(service, eventWithBody)

	case "reverse-geocode":
		return application.ReverseGeocodeOperation(service, eventWithBody)

	case "validate-postal":
		return application.ValidatePostalOperation(service, eventWithBody)

	case "validate-municipality":
		return application.ValidateMunicipalityOperation(service, eventWithBody)

	case "autocomplete-postal":
		return application.AutocompletePostalOperation(service, eventWithBody)

	case "autocomplete-municipality":
		return application.AutocompleteMunicipalityOperation(service, eventWithBody)

	default:
		errorBody, _ := json.Marshal(map[string]interface{}{
			"success": false,
			"error":   "Operation '" + body.Operation + "' not supported. Supported operations: geocode-by-postal, reverse-geocode, validate-postal, validate-municipality, autocomplete-postal, autocomplete-municipality",
		})
		return domain.LambdaResponse{
			StatusCode: 400,
			Body:       string(errorBody),
		}, nil
	}
}

// StartLambda starts the Lambda function handler.
//
// This is a convenience function that wraps lambda.Start and initializes
// the AWS Lambda runtime. The handler will process incoming events and
// route them to the appropriate geocoding operations.
//
// This function blocks indefinitely, processing Lambda invocations as they arrive.
func StartLambda() {
	lambda.Start(Handler)
}

