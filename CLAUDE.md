# pricofy-geocode-es - Spanish Postal Code Geocoding Service

**Repository:** https://github.com/Pricofy/pricofy-geocode-es  
**Purpose:** Spanish postal code geocoding, validation, and autocompletion microservice  
**Tech Stack:** Go 1.24+, AWS Lambda (PROVIDED_AL2023), Static Postal DB (embedded)  
**Deployment:** Single Lambda function via CDK  
**Invocation:** Lambda SDK (no API Gateway, no HTTP)  
**Version:** 1.0.0 (First production release)  

**Global Context:** See [pricofy-docs/CLAUDE.md](https://github.com/cnebrera/pricofy-docs/blob/main/CLAUDE.md)

**Quick Links:**
- [README](./README.md) - Overview and quick start
- [API Documentation](./api/README.md) - Lambda invocation contract
- [Postal Codes DB](./docs/POSTAL_CODES.md) - Database documentation

---

## ⚠️ Critical: This is NOT a REST API

**This service is invoked via AWS Lambda SDK**, not HTTP/API Gateway.

**Why?**
- ✅ **Single cold start** - All 6 operations share one Lambda instance and in-memory data (11,150 postal codes)
- ✅ **10-20ms faster** - No API Gateway overhead
- ✅ **$0 cost** - No API Gateway charges
- ✅ **Simpler** - One Lambda, one deployment

**Invocation:**
```typescript
const lambda = new Lambda();
const result = await lambda.invoke({
  FunctionName: 'pricofy-geocode-es-dev',
  Payload: JSON.stringify({
    operation: 'geocode-by-postal',
    postalCode: '28001'
  })
}).promise();
```

**OpenAPI exists** to document the payload contract, not HTTP endpoints.

---

## What This Does

Provides Spanish-specific geocoding services for Pricofy:
- **Postal Code Geocoding:** Convert Spanish postal codes to coordinates (11,150 entries)
- **Municipality Geocoding:** Convert Spanish municipality names to coordinates
- **Reverse Geocoding:** Find nearest postal code from GPS coordinates (Haversine distance)
- **Postal Code Validation:** Validate if a postal code exists in Spain
- **Municipality Validation:** Validate if a municipality name exists
- **Postal Code Autocompletion:** Autocomplete postal codes as user types
- **Municipality Autocompletion:** Autocomplete municipality names as user types

**Invoked by:** pricofy-location-service (orchestrator) via Lambda invoke  
**Security:** Private Lambda function with IAM authentication only  
**Performance:** <1ms geocoding (postal code), ~10-20ms reverse geocoding, <5ms autocompletion  
**Cold Start:** ~200ms (3-5x faster than Node.js ~500ms)  
**Memory:** 128MB (50% reduction from Node.js 256MB)

---

## Key Features

### 1. Hexagonal Architecture (Ports & Adapters)

Clean separation of concerns across three layers:

```
┌─────────────────────────────────────────────────────────────────┐
│                    Entry Point (Handler)                         │
│                    cmd/lambda/main.go                             │
│                    → handler.Handler()                            │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                   Operations Layer                               │
│  internal/application/operations.go                             │
│  geocode-by-postal, reverse-geocode, validate-postal,           │
│  validate-municipality, autocomplete-postal, autocomplete-municipality│
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                   Application Layer (Service)                    │
│  internal/application/service.go                                │
│  PostalCodeService - Orchestrates geocoding operations          │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│              Infrastructure Layer (Provider)                     │
│  internal/infrastructure/provider/postal_provider.go             │
│  PostalCodeProvider - Static Spanish postal codes (embedded)   │
│                                                                  │
│  Shared Services:                                               │
│  - Logger (structured CloudWatch logs)                          │
│  - Custom Error Types (domain errors)                           │
└─────────────────────────────────────────────────────────────────┘
```

**Benefits:**
- Core business logic isolated from infrastructure
- Easy to test with real data (no mocks needed)
- Add new operations without modifying existing code
- Single Lambda function with operation-based routing

### 2. Single Lambda Function with 6 Operations

**Function:** `pricofy-geocode-es-{env}`  
**Handler:** `bootstrap` (Go binary entry point)  
**Runtime:** `PROVIDED_AL2023` (Go custom runtime)  
**Memory:** 128MB (optimized for Go, reduced from 256MB Node.js)  
**Routing:** Based on `operation` field in request body

| Operation | Purpose | Input | Output | Latency |
|-----------|---------|-------|--------|---------|
| **geocode-by-postal** | Postal Code → Coords | `postalCode` or `municipality` | Coordinates + metadata | <1ms |
| **reverse-geocode** | Coords → Postal Code | `lat`, `lon` | Nearest postal code + distance | ~10-20ms |
| **validate-postal** | Check if postal code is valid | `postalCode` | `valid: true/false` | <1ms |
| **validate-municipality** | Check if municipality is valid | `municipality` | `valid: true/false` | <5ms |
| **autocomplete-postal** | Autocomplete postal codes | `query`, `limit` | List of matching postal codes | <5ms |
| **autocomplete-municipality** | Autocomplete municipalities | `query`, `limit` | List of matching municipalities | <5ms |

**All operations are PRIVATE** (no public URLs, no API Gateway).  
Only invokable by **pricofy-location-service** via IAM role.

### 3. Static Spanish Postal Codes Database

**Source:** GeoNames (11,150 unique Spanish postal codes)

**Structure:**
```json
{
  "28001": {
    "lat": 40.4168,
    "lon": -3.7038,
    "municipality": "Madrid",
    "province": "Madrid"
  }
}
```

**In-Memory Indices for Performance:**
- **Postal Code Map:** `map[string]PostalData` - O(1) lookup by postal code
- **Municipality Index:** `map[string][]postalEntry` - O(1) lookup by municipality name
- **Municipality Set:** `map[string]bool` - O(1) validation
- **Sorted Postal Codes:** `[]string` - Binary search for autocompletion
- **All Postal Codes Array:** Linear search for reverse geocoding (Haversine)

**Performance:**
- Postal code lookup: O(1) - <1ms
- Municipality search: O(1) - <1ms
- Autocompletion: O(log n) - <5ms
- Reverse geocoding: O(n) - ~10-20ms (Haversine distance calculation)

**Benefits:**
- ✅ Unlimited requests (no rate limits)
- ✅ Zero external API dependencies
- ✅ Sub-millisecond latency
- ✅ No cost per request
- ✅ Offline operation

### 4. Structured Logging & Observability

**Logger Utility:** Structured JSON logs for CloudWatch Insights

```go
logger.Info("PostalCodeService", "Geocoding completed", map[string]interface{}{
    "postalCode": "28001",
    "latency":    5,
    "source":     "postal_code",
})
```

**Log Levels:**
- DEBUG - Detailed debugging information
- INFO - Important state changes
- WARN - Recoverable issues
- ERROR - Failures with full stack traces

**CloudWatch Insights Queries:**
```sql
-- Find all errors in last hour
fields @timestamp, component, message, error
| filter level = "ERROR"
| sort @timestamp desc

-- Track geocoding success rate
fields @timestamp, component
| filter component = "PostalCodeService" and message like /successful/
| stats count() as successes by bin(5m)
```

### 5. Custom Error Hierarchy

**Type-safe error handling:**

```go
LocationError (base struct)
├── InvalidCoordinatesError       // Invalid lat/lon values
├── PostalCodeNotFoundError       // Postal code not in database
└── ValidationError               // Request validation failures
```

All errors implement the `error` interface and include timestamps for debugging.

**Benefits:**
- Better error categorization
- Easier debugging with timestamps
- Consistent error responses
- Type-safe error handling

---

## Security Model

### Private Lambda Function

```
Frontend (Flutter)
    ↓ [HTTPS + X-Api-Key]
API Gateway (pricofy-api)
    ↓ [Lambda Invoke + IAM Role]
pricofy-location-service (ORCHESTRATOR)
    ↓ [Lambda Invoke + IAM Role]
pricofy-geocode-es (COUNTRY-SPECIFIC)
    ├─ geocode-by-postal
    ├─ reverse-geocode
    ├─ validate-postal
    ├─ validate-municipality
    ├─ autocomplete-postal
    └─ autocomplete-municipality
```

**Why this architecture:**
- ✅ Country-specific logic isolated in dedicated service
- ✅ Location service orchestrates global + country-specific operations
- ✅ Easy to add new countries (e.g., pricofy-geocode-fr)
- ✅ pricofy-api controls ACL and authorization

### IAM Permissions

**Geocode ES Function:**
- No public endpoints (no API Gateway, no Function URLs)
- Only invokable via `lambda:InvokeFunction` permission
- pricofy-location-service has invoke permission

---

## Operations (Lambda Functions)

### Request Format (All Operations)

**Common structure:**
```json
{
  "body": "{\"operation\":\"geocode-by-postal\",\"country\":\"ES\",\"postalCode\":\"28001\"}"
}
```

### 1. geocode-by-postal

**Purpose:** Convert Spanish postal code or municipality to coordinates

**Input (by postal code):**
```json
{
  "operation": "geocode-by-postal",
  "country": "ES",
  "postalCode": "28001"
}
```

**Input (by municipality):**
```json
{
  "operation": "geocode-by-postal",
  "country": "ES",
  "municipality": "Madrid"
}
```

**Output:**
```json
{
  "statusCode": 200,
  "body": "{\"success\":true,\"coords\":{\"lat\":40.4168,\"lon\":-3.7038},\"municipality\":\"Madrid\",\"province\":\"Madrid\",\"postalCode\":\"28001\",\"source\":\"postal_code\"}"
}
```

**Features:**
- Static database of 11,150 Spanish postal codes (embedded in binary)
- Unlimited requests (no rate limit)
- <1ms latency for postal code lookup (O(1))
- <1ms latency for municipality search (O(1) with in-memory index)
- No external API dependencies
- Memory: 128MB (Go implementation, 50% reduction from Node.js)

**Accuracy:**
- Postal code: Centroid of postal code area (±500m typical)
- Municipality: Centroid of municipality (±2-5km)

**Source Field:**
- `postal_code` - Looked up by postal code (more precise)
- `municipality` - Looked up by municipality name (less precise)

### 2. reverse-geocode

**Purpose:** Find nearest Spanish postal code from GPS coordinates

**Input:**
```json
{
  "operation": "reverse-geocode",
  "country": "ES",
  "lat": 40.4168,
  "lon": -3.7038
}
```

**Output:**
```json
{
  "statusCode": 200,
  "body": "{\"success\":true,\"city\":\"Madrid\",\"postalCode\":\"28001\",\"province\":\"Madrid\",\"country\":\"España\",\"coords\":{\"lat\":40.4168,\"lon\":-3.7038},\"distance\":0.142}"
}
```

**Features:**
- Haversine distance calculation to find nearest postal code
- Searches all 11,150 postal codes (optimized linear scan)
- ~10-20ms latency
- Returns distance to nearest postal code centroid (km, rounded to 3 decimals)
- Memory: 128MB (Go implementation)

**Accuracy:**
- Distance field indicates precision (smaller = more accurate)
- Typical: ±500m to 2km (depends on postal code density)
- Rural areas: ±5-10km (larger postal code areas)

### 3. validate-postal

**Purpose:** Validate if a Spanish postal code exists

**Input:**
```json
{
  "operation": "validate-postal",
  "country": "ES",
  "postalCode": "28001"
}
```

**Output:**
```json
{
  "statusCode": 200,
  "body": "{\"success\":true,\"valid\":true,\"postalCode\":\"28001\"}"
}
```

**Features:**
- O(1) lookup in postal code index
- <1ms latency
- Returns `valid: true` if postal code exists, `valid: false` otherwise

### 4. validate-municipality

**Purpose:** Validate if a Spanish municipality name exists

**Input:**
```json
{
  "operation": "validate-municipality",
  "country": "ES",
  "municipality": "Madrid"
}
```

**Output:**
```json
{
  "statusCode": 200,
  "body": "{\"success\":true,\"valid\":true,\"municipality\":\"Madrid\"}"
}
```

**Features:**
- O(1) lookup in municipality index (case-insensitive)
- <1ms latency
- Returns `valid: true` if municipality exists, `valid: false` otherwise

### 5. autocomplete-postal

**Purpose:** Autocomplete Spanish postal codes as user types

**Input:**
```json
{
  "operation": "autocomplete-postal",
  "country": "ES",
  "query": "280",
  "limit": 10
}
```

**Output:**
```json
{
  "statusCode": 200,
  "body": "{\"success\":true,\"results\":[{\"value\":\"28001\",\"label\":\"28001 - Madrid (Madrid)\"},{\"value\":\"28002\",\"label\":\"28002 - Madrid (Madrid)\"}]}"
}
```

**Features:**
- Binary search in sorted postal code index
- <5ms latency
- Returns up to `limit` results (default: 10, max: 50)
- Case-insensitive prefix matching

### 6. autocomplete-municipality

**Purpose:** Autocomplete Spanish municipality names as user types

**Input:**
```json
{
  "operation": "autocomplete-municipality",
  "country": "ES",
  "query": "mad",
  "limit": 10
}
```

**Output:**
```json
{
  "statusCode": 200,
  "body": "{\"success\":true,\"results\":[{\"value\":\"Madrid\",\"label\":\"Madrid (Madrid)\"},{\"value\":\"Madarcos\",\"label\":\"Madarcos (Madrid)\"}]}"
}
```

**Features:**
- Binary search in sorted municipality index
- <5ms latency
- Returns up to `limit` results (default: 10, max: 50)
- Case-insensitive prefix matching

---

## Infrastructure (CDK)

### Single CDK Stack

**Purpose:** Spanish geocoding Lambda function

**Resources:**
- Lambda function: `pricofy-geocode-es-{env}`
- IAM role: Minimal permissions (CloudWatch Logs only)

**Deployment Order:** Deploy after pricofy-location-service

**Exports (CloudFormation):**
- `Pricofy-GeocodeEsArn-{env}` - ARN of geocode-es function
- `Pricofy-GeocodeEsName-{env}` - Name of geocode-es function

---

## Deployment

### Via Makefile (Recommended)

**Full deployment (clean + test + deploy):**
```bash
make deploy ENV=dev
```

**Quick deployment (skip tests):**
```bash
make deploy-quick ENV=dev
```

**What it does:**
1. `make clean` - Remove build artifacts
2. `make install` - Install Go dependencies + CDK dependencies
3. `make build` - Compile Go to Linux AMD64 binary (dist/bootstrap)
4. `make test` - Run Go tests (if using `make deploy`)
5. Deploy Geocode ES Stack via CDK

**Available Makefile targets:**
```bash
make help              # Show all available commands
make install           # Install Go + CDK dependencies
make build             # Compile Go to Linux AMD64 binary
make test              # Run Go tests with coverage
make lint              # Run golangci-lint
make clean             # Clean build artifacts
make verify ENV=dev    # Verify deployment prerequisites
make deploy ENV=dev    # Safe deployment (clean + test + deploy)
make deploy-quick ENV=dev  # Quick deploy (skip tests)
make destroy-dev       # Destroy dev environment
make destroy-prod      # Destroy prod environment
```

### Via GitHub Actions (Auto-deploy)

**On push to develop branch:**
- Automatic deployment to DEV environment

**Manual dispatch:**
- Deploy to DEV or PROD via GitHub UI

**Workflow:** `.github/workflows/deploy.yml`

**Features:**
- Pre-deployment verification (AWS CLI, CDK bootstrap)
- OIDC authentication (no long-lived credentials)
- Comprehensive logging
- Deployment status notifications

**Required GitHub Secrets:**
- `AWS_ACCOUNT_ID_DEV` - AWS account ID for dev environment
- `AWS_ACCOUNT_ID_PROD` - AWS account ID for prod environment
- `AWS_ROLE_ARN_DEV` - IAM role ARN for dev deployment
- `AWS_ROLE_ARN_PROD` - IAM role ARN for prod deployment

### First-Time Setup

**Prerequisites:**
1. AWS CLI configured with valid credentials
2. CDK bootstrapped in target account/region
3. GitHub OIDC provider configured (for GitHub Actions)

**Setup Steps:**

```bash
# 1. Install dependencies
make install

# 2. Verify deployment prerequisites
make verify ENV=dev

# 3. Deploy
make deploy ENV=dev
```

---

## Testing

### Unit Tests

**Run tests:**
```bash
make test              # Run all tests with coverage
go test ./test/... -v  # Run tests with verbose output
go test ./test/... -cover  # Run tests with coverage report
```

**Test Coverage:**
- Provider: 100% (data access, indices, geocoding operations)
- Service: 100% (business logic, validation, error handling)
- Handler: Integration tests for Lambda event routing
- Overall: 100% coverage target

**Test Structure:**
```
test/
├── unit/
│   ├── provider_test.go    # PostalCodeProvider tests
│   └── service_test.go     # PostalCodeService tests
└── integration/
    └── handler_test.go     # Lambda handler integration tests
```

### Manual Testing (Lambda Invoke)

```bash
# Test geocode-by-postal (postal code)
aws lambda invoke \
  --function-name pricofy-geocode-es-dev \
  --payload '{"body":"{\"operation\":\"geocode-by-postal\",\"country\":\"ES\",\"postalCode\":\"28001\"}"}' \
  response.json

cat response.json

# Test geocode-by-postal (municipality)
aws lambda invoke \
  --function-name pricofy-geocode-es-dev \
  --payload '{"body":"{\"operation\":\"geocode-by-postal\",\"country\":\"ES\",\"municipality\":\"Madrid\"}"}' \
  response.json

cat response.json

# Test reverse-geocode
aws lambda invoke \
  --function-name pricofy-geocode-es-dev \
  --payload '{"body":"{\"operation\":\"reverse-geocode\",\"country\":\"ES\",\"lat\":40.4168,\"lon\":-3.7038}"}' \
  response.json

cat response.json

# Test validate-postal
aws lambda invoke \
  --function-name pricofy-geocode-es-dev \
  --payload '{"body":"{\"operation\":\"validate-postal\",\"country\":\"ES\",\"postalCode\":\"28001\"}"}' \
  response.json

cat response.json

# Test validate-municipality
aws lambda invoke \
  --function-name pricofy-geocode-es-dev \
  --payload '{"body":"{\"operation\":\"validate-municipality\",\"country\":\"ES\",\"municipality\":\"Madrid\"}"}' \
  response.json

cat response.json

# Test autocomplete-postal
aws lambda invoke \
  --function-name pricofy-geocode-es-dev \
  --payload '{"body":"{\"operation\":\"autocomplete-postal\",\"country\":\"ES\",\"query\":\"280\",\"limit\":10}"}' \
  response.json

cat response.json

# Test autocomplete-municipality
aws lambda invoke \
  --function-name pricofy-geocode-es-dev \
  --payload '{"body":"{\"operation\":\"autocomplete-municipality\",\"country\":\"ES\",\"query\":\"mad\",\"limit\":10}"}' \
  response.json

cat response.json
```

---

## Integration with pricofy-location-service

pricofy-location-service **delegates Spanish operations** to this function:

```typescript
// pricofy-location-service/src/handlers/geocode.ts (orchestrator)
export async function handler(event: APIGatewayProxyEvent) {
  const { country, operation } = JSON.parse(event.body);
  
  // Route to country-specific Lambda
  if (country === 'ES') {
    const result = await lambda.invoke({
      FunctionName: process.env.GEOCODE_ES_FUNCTION_NAME,
      InvocationType: 'RequestResponse',
      Payload: JSON.stringify(event),
    }).promise();
    
    return JSON.parse(result.Payload as string);
  }
  
  // ... other countries or global operations
}
```

**IAM Permissions:**
- pricofy-location-service Lambda execution role has `lambda:InvokeFunction` on this function ARN
- Configured in pricofy-infra CDK

---

## Monitoring

### CloudWatch Metrics

**Key Metrics:**
- Invocations (total requests)
- Duration (execution time: p50, p95, p99)
- Errors (failed invocations)
- Throttles (rate limiting)
- Concurrent Executions

### CloudWatch Logs

**Structured JSON logs:**

```bash
# geocode-es logs
aws logs tail /aws/lambda/pricofy-geocode-es-dev --follow
```

**CloudWatch Insights Queries:**

```sql
-- Find all errors in last hour
fields @timestamp, component, message, error
| filter level = "ERROR"
| sort @timestamp desc
| limit 100

-- Track geocoding success rate
fields @timestamp, component
| filter component = "PostalCodeService" and message like /successful/
| stats count() as successes by bin(5m)

-- Average geocoding latency
fields @timestamp, @duration
| filter @message like /Geocoding completed/
| stats avg(@duration) as avg_ms by bin(1h)

-- Most requested postal codes
fields component, metadata.postalCode
| filter component = "PostalCodeService" and metadata.postalCode != null
| stats count() as requests by metadata.postalCode
| sort requests desc
| limit 20
```

---

## Cost Analysis

### Monthly Costs (Dev Environment, ~500 requests/day)

| Resource | Usage | Cost |
|----------|-------|------|
| **Lambda Invocations** | 15K/month | $0 (free tier: 1M) |
| **Lambda Duration** | 50K GB-seconds | $0 (free tier: 400K) |
| **CloudWatch Logs** | 0.5 GB | $0.25 |
| **Total** | | **~$0.25/month** |

**Production:** Similar costs (within free tier for moderate traffic)

**Cost Optimization:**
- Static Postal DB: Zero cost per request
- No external API calls: Zero usage-based costs
- Lambda free tier: 1M requests/month

---

## Common Issues

### 1. "Postal code not found"

**Cause:** Postal code not in database (only Spanish postal codes supported)

**Solution:**
- Verify postal code is valid Spanish format (5 digits, 01000-52999)
- Check `src/resources/postal-codes-es.json` for coverage
- For non-Spanish postal codes, use different geocoding service

### 2. Cold start too slow (>100ms)

**Current:** ~50-100ms (loading 11,150 postal codes into memory)

**Solution: Lambda Warmup (Implemented)**
- CloudWatch Events trigger warmup every 5 minutes
- Self-invokes with concurrency=2 to maintain 3 warm instances
- Warmup detected before operation routing, returns immediately
- Files: `cmd/warmup.go`, `cmd/main.go` (warmup detection)

---

## Maintenance

### Update Postal Codes Database

**Frequency:** As needed (GeoNames updates quarterly)

**Process:**
```bash
# 1. Download latest GeoNames data
cd scripts
curl -O http://download.geonames.org/export/zip/ES.zip
unzip ES.zip

# 2. Regenerate postal codes JSON
node convert-postal-codes.js

# 3. Commit updated file
git add src/resources/postal-codes-es.json
git commit -m "chore: update Spanish postal codes database"

# 4. Deploy
make deploy ENV=dev
```

### Code Updates

**Via GitHub Actions (recommended):**
- Push to `develop` branch → Auto-deploy to DEV
- Manual dispatch → Deploy to PROD

**Via Makefile (manual):**
```bash
make deploy ENV=dev
```

---

## Code Quality & Refactoring (November 2024)

### Architecture Improvements

**1. Hexagonal Architecture** - Clean separation of handlers, operations, services, and providers
**2. Logger Utility** - Structured JSON logging for CloudWatch Insights
**3. Custom Error Hierarchy** - 5 error types for better categorization
**4. Constants Module** - Centralized configuration values
**5. Provider Interfaces** - Dependency inversion for testability
**6. In-Memory Indices** - Optimized data structures for fast lookups and autocompletion

### Code Quality Metrics

**Test Coverage:**
- Overall: 100%
- Handlers: 100%
- Operations: 100%
- Providers: 100%
- Services: 100%
- Utils: 100%

---

## Related Projects

- **pricofy-location-service** - Main location orchestrator (invokes this service)
- **pricofy-api** - Main API Gateway (invokes location-service)
- **pricofy-frontend** - Flutter frontend (consumes via pricofy-api)
- **pricofy-infra** - Shared infrastructure (IAM roles, OIDC, CDK bootstrap)

---

**Last Updated:** November 18, 2025  
**Version:** 1.0.0  
**Status:** ✅ Ready for development

**Key Features:**
- ✅ Spanish postal code geocoding (11,150 entries)
- ✅ Reverse geocoding with Haversine distance
- ✅ Postal code and municipality validation
- ✅ Autocomplete for postal codes and municipalities
- ✅ In-memory indices for O(1) lookups
- ✅ Implemented hexagonal architecture
- ✅ Structured logging and custom errors
- ✅ 100% test coverage
