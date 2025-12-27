// Package main contains the Lambda warmup handler for preventing cold starts.
// CloudWatch Events trigger this handler periodically to keep Lambda instances warm.
package main

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	lambdasdk "github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/pricofy/geocode-pt/internal/domain"
)

const (
	// WarmupSource identifies warmup events from CloudWatch
	WarmupSource = "warmup"

	// WarmupDelay ensures instances overlap to create true concurrency
	WarmupDelay = 75 * time.Millisecond
)

// WarmupEvent represents the CloudWatch Event payload for warmup
type WarmupEvent struct {
	Source      string `json:"source"`
	Concurrency int    `json:"concurrency"`
}

// WarmupResponse is the response returned by warmup operations
type WarmupResponse struct {
	Status          string `json:"status"`
	InstancesWarmed int    `json:"instancesWarmed"`
}

// IsWarmupEvent checks if the event is a warmup event
func IsWarmupEvent(event domain.LambdaEvent) (*WarmupEvent, bool) {
	// Try to parse body as JSON to check for warmup source
	if event.Body == "" {
		return nil, false
	}

	var bodyMap map[string]interface{}
	if err := json.Unmarshal([]byte(event.Body), &bodyMap); err != nil {
		return nil, false
	}

	source, ok := bodyMap["source"].(string)
	if !ok || source != WarmupSource {
		return nil, false
	}

	warmup := &WarmupEvent{
		Source:      source,
		Concurrency: 0,
	}

	// Parse concurrency (optional, defaults to 0)
	if concurrency, ok := bodyMap["concurrency"].(float64); ok {
		warmup.Concurrency = int(concurrency)
	}

	return warmup, true
}

// HandleWarmup processes a warmup event and optionally self-invokes
// to maintain multiple warm instances.
func HandleWarmup(ctx context.Context, warmup *WarmupEvent) (domain.LambdaResponse, error) {
	instancesWarmed := 1 // This instance counts as 1

	if warmup.Concurrency > 0 {
		if err := selfInvoke(ctx, warmup.Concurrency); err == nil {
			instancesWarmed += warmup.Concurrency
		}
	}

	// Brief delay to ensure instances overlap
	time.Sleep(WarmupDelay)

	body, _ := json.Marshal(WarmupResponse{
		Status:          "warm",
		InstancesWarmed: instancesWarmed,
	})

	return domain.LambdaResponse{
		StatusCode: 200,
		Body:       string(body),
	}, nil
}

// selfInvoke invokes this Lambda function N times asynchronously
// to create additional warm instances.
func selfInvoke(ctx context.Context, count int) error {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}

	client := lambdasdk.NewFromConfig(cfg)
	functionName := os.Getenv("AWS_LAMBDA_FUNCTION_NAME")

	// Payload for child invocations (concurrency=0 to prevent infinite loop)
	warmupBody, _ := json.Marshal(WarmupEvent{
		Source:      WarmupSource,
		Concurrency: 0, // Critical: prevent recursive invocation
	})
	payload, _ := json.Marshal(domain.LambdaEvent{
		Body: string(warmupBody),
	})

	// Invoke in parallel
	var wg sync.WaitGroup
	var invokeErr error
	var errMu sync.Mutex

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			_, err := client.Invoke(ctx, &lambdasdk.InvokeInput{
				FunctionName:   aws.String(functionName),
				InvocationType: types.InvocationTypeEvent, // Async invocation
				Payload:        payload,
			})

			if err != nil {
				errMu.Lock()
				if invokeErr == nil {
					invokeErr = err
				}
				errMu.Unlock()
			}
		}()
	}

	wg.Wait()
	return invokeErr
}
