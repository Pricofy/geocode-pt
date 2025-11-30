# Implementation Guide: Adding New Countries to Geocode Service

This guide provides step-by-step instructions for adding support for new countries to the Pricofy geocoding service. The process has been validated with France and can be replicated for any country with postal code data.

## Prerequisites

### Required Tools
- **Go 1.24+**: For compilation and testing
- **Node.js 20+**: For CDK infrastructure and E2E tests
- **AWS CLI**: Configured with appropriate credentials
- **CDK CLI**: For infrastructure deployment

### Required Data
- **GeoNames postal code dataset**: Available at `https://download.geonames.org/export/zip/{COUNTRY_CODE}.zip`
- **Country code**: 2-letter ISO code (FR, DE, IT, PT, etc.)

## Step-by-Step Implementation

### Phase 1: Data Preparation

#### 1.1 Download and Convert Postal Code Data

```bash
# Create data directory
mkdir -p data

# Download country dataset (example: Germany)
curl -O https://download.geonames.org/export/zip/DE.zip
unzip DE.zip

# Convert to JSON format
node scripts/convert-geonames-to-json.js DE.txt DE > data/postal-codes-de.json
```

#### 1.2 Validate Converted Data

```bash
# Run validation script
node scripts/validate-postal-data.js data/postal-codes-de.json DE

# Expected output: "Overall Assessment: EXCELLENT"
```

#### 1.3 Check Data Statistics

```bash
# Get file size and basic stats
ls -lh data/postal-codes-de.json
wc -l data/postal-codes-de.json
```

### Phase 2: Clone and Adapt Service

#### 2.1 Clone Base Service

```bash
# Clone from Spain service (template)
cp -r pricofy-geocode-es pricofy-geocode-de

# Enter new service directory
cd pricofy-geocode-de
```

#### 2.2 Update Go Module

```bash
# Update go.mod
sed -i '' 's|github.com/pricofy/geocode-es|github.com/pricofy/geocode-de|g' go.mod

# Update sonar-project.properties
sed -i '' 's|pricofy-geocode-es|pricofy-geocode-de|g' sonar-project.properties
sed -i '' 's|pricofy_pricofy-geocode-es|pricofy_pricofy-geocode-de|g' sonar-project.properties
```

#### 2.3 Copy Country Data

```bash
# Copy the converted postal code data
cp ../data/postal-codes-de.json internal/infrastructure/provider/postal-codes-de.json
```

#### 2.4 Update All Go Imports

```bash
# Update all import paths in Go files
find . -name "*.go" -type f -exec sed -i '' 's|github\.com/pricofy/geocode-es|github.com/pricofy/geocode-de|g' {} \;
```

#### 2.5 Update Postal Code Provider

```go
// In internal/infrastructure/provider/postal_provider.go
//go:embed postal-codes-de.json  // Change from postal-codes-es.json
var postalCodesJSON []byte
```

#### 2.6 Update Regex Patterns

```go
// In internal/domain/constants.go
const PostalCodeRegexPattern = `^\d{5}$`  // Germany: 5 digits (same as Spain)

// Special cases:
// France: `^\d{5}(\s+[A-Z\s]+)?$` (allows CEDEX)
// Portugal: `^\d{4}-\d{3}$|^\d{7}$` (with or without hyphen)
```

#### 2.7 Update Tests

**Provider Tests (`internal/infrastructure/provider/postal_provider_test.go`):**
```go
// Update test data to use real postal codes from the new country
{
    name:        "valid postal code - Berlin",  // Instead of Madrid
    postalCode:  "10115",                       // Berlin postal code
    wantSuccess: true,
    wantErr:     false,
}
```

**Application Tests (`internal/application/service_test.go`):**
- Update postal codes and municipalities in test cases
- Ensure coordinates are within the country's bounds

#### 2.8 Update Country Name in Reverse Geocoding

```go
// In internal/application/service.go
Country: "Germany",  // Instead of "España"
```

#### 2.9 Update Comments and Documentation

```go
// Update all comments mentioning "Spanish" to the new country
// Examples:
// "Spanish postal code geocoding service" → "German postal code geocoding service"
// "Madrid" → "Berlin"
// "Barcelona" → "Munich"
```

### Phase 3: Infrastructure (CDK) Updates

#### 3.1 Rename CDK Stack File

```bash
# Rename CDK stack file
mv infrastructure/lib/geocode-es-stack.ts infrastructure/lib/geocode-de-stack.ts

# Rename test file
mv infrastructure/test/geocode-es-stack.test.ts infrastructure/test/geocode-de-stack.test.ts
```

#### 3.2 Update CDK Stack Class

```typescript
// In infrastructure/lib/geocode-de-stack.ts
export class GeocodeDeStack extends cdk.Stack {  // Instead of GeocodeEsStack
  // Update all references:
  // FunctionName: 'pricofy-geocode-es' → 'pricofy-geocode-de'
  // Log group: /aws/lambda/pricofy-geocode-es → /aws/lambda/pricofy-geocode-de
  // Exports: Pricofy-GeocodeEsArn → Pricofy-GeocodeDeArn
  // Description: "Spanish" → "German"
}
```

#### 3.3 Update CDK Tests

```typescript
// In infrastructure/test/geocode-de-stack.test.ts
import { GeocodeDeStack } from '../lib/geocode-de-stack';

// Update all test expectations to match new names
template.hasResourceProperties('AWS::Lambda::Function', {
  FunctionName: 'pricofy-geocode-de',  // Updated name
});
```

### Phase 4: Build System Updates

#### 4.1 Update Makefile Variables

```makefile
# In Makefile
BINARY = pricofy-geocode-de          # Updated
SERVICE_NAME = pricofy-geocode-de    # Updated
STACK_SERVICE = PricofyGeocodeDeStack # Updated
LAMBDA_GEOCODE = pricofy-geocode-de  # Updated
```

#### 4.2 Update GitHub Actions

**Deploy Workflow (`deploy.yml`):**
```yaml
# Update workflow name and references
name: Deploy Geocode DE  # Instead of "Deploy Geocode ES"
```

**Quality Workflow (`quality.yml`):**
- Update workflow name and job references

**E2E Tests (`e2e-tests.yml`):**
- Update Lambda function names in test configurations

### Phase 5: E2E Test Updates

#### 5.1 Update Test Configuration

```typescript
// In test/e2e/src/config.ts
export const CONFIG = {
  functionName: 'pricofy-geocode-de',  // Updated
  // ... other config
};
```

#### 5.2 Update Test Cases

```typescript
// In test/e2e/src/geocode-de.test.ts (renamed from geocode-es.test.ts)

// Update all test data to use real postal codes from Germany
const testPostalCode = '10115';  // Berlin instead of 28001 (Madrid)
const testMunicipality = 'Berlin';  // Instead of Madrid
const testCoordinates = { lat: 52.5172, lon: 13.4149 };  // Berlin coordinates
```

### Phase 6: Documentation Updates

#### 6.1 Update README.md

- Service description
- API examples with country-specific postal codes
- Coverage information

#### 6.2 Update docs/POSTAL_CODES.md

- Update statistics (total codes, size, ranges)
- Update data source information
- Document country-specific format quirks

#### 6.3 Update docs/CONTRACTS.md

- Update API examples with real postal codes
- Update response examples with country-specific data

#### 6.4 Reset CHANGELOG.md

```markdown
# Changelog

## [1.0.0] - 2025-11-XX

### Added
- Initial release of German postal code geocoding service
- Support for 10,813 German postal codes
- All standard operations: geocode, reverse-geocode, validate, autocomplete
```

### Phase 7: Testing and Validation

#### 7.1 Run All Tests

```bash
# Install dependencies
go mod tidy
cd infrastructure && npm install

# Run all tests
make test        # Go + CDK tests
make test-e2e    # E2E tests (requires deployed Lambda)
```

#### 7.2 Build and Verify

```bash
# Build for Lambda
make build

# Verify binary
ls -lh dist/bootstrap
file dist/bootstrap
```

### Phase 8: Deployment

#### 8.1 Verify Prerequisites

```bash
# Check AWS configuration
make verify ENV=dev
```

#### 8.2 Deploy to Development

```bash
# Safe deployment (includes tests)
make deploy ENV=dev
```

#### 8.3 Test Deployed Service

```bash
# Test basic functionality
make test-geocode ENV=dev

# Run E2E tests against deployed service
make test-e2e ENV=dev
```

#### 8.4 Deploy to Production

```bash
# Deploy to production when ready
make deploy ENV=prod
```

## Country-Specific Considerations

### Memory Requirements

| Country | Codes | JSON Size | Lambda Memory | Notes |
|---------|-------|-----------|---------------|-------|
| France | 20,414 | 2.3MB | 128MB | Standard |
| Germany | 10,813 | 1.3MB | 128MB | Standard |
| Italy | ~4,484 | ~1MB | 128MB | Standard |
| Portugal | 197,772 | 23MB | 256MB | ⚠️ Requires more memory |

### Postal Code Formats

| Country | Format | Regex | Notes |
|---------|--------|-------|-------|
| Spain | 28001 | `^\d{5}$` | Standard 5 digits |
| France | 75001 | `^\d{5}(\s+[A-Z\s]+)?$` | Allows CEDEX suffixes |
| Germany | 10115 | `^\d{5}$` | Standard 5 digits |
| Italy | 00118 | `^\d{5}$` | Standard 5 digits |
| Portugal | 1000-205 | `^\d{4}-\d{3}$|^\d{7}$` | With/without hyphen |

### Administrative Divisions

| Country | Province Equivalent | Notes |
|---------|---------------------|-------|
| Spain | Province | `admin name2` |
| France | Département | `admin name2` |
| Germany | State | `admin name1` |
| Italy | Province | `admin name2` |
| Portugal | District | `admin name1` |

## Automation Script

For faster implementation, use this automation script:

```bash
#!/bin/bash
# automate-country-setup.sh

COUNTRY_CODE=$1  # e.g., "DE"
COUNTRY_NAME=$2  # e.g., "Germany"

if [ -z "$COUNTRY_CODE" ] || [ -z "$COUNTRY_NAME" ]; then
    echo "Usage: $0 <country_code> <country_name>"
    echo "Example: $0 DE Germany"
    exit 1
fi

echo "Setting up geocoding service for $COUNTRY_NAME ($COUNTRY_CODE)..."

# Add automation commands here
# This would include all the steps above
```

## Troubleshooting

### Common Issues

1. **Import errors**: Ensure all Go imports are updated consistently
2. **Test failures**: Update test data with real postal codes from the new country
3. **CDK deployment failures**: Check Lambda function names and IAM permissions
4. **E2E test failures**: Verify function names in test configuration

### Validation Checklist

- [ ] All Go tests pass (`make test`)
- [ ] CDK tests pass (`make test-cdk`)
- [ ] Binary builds successfully (`make build`)
- [ ] E2E tests pass (`make test-e2e`)
- [ ] Manual testing works (`make test-geocode`)
- [ ] Performance acceptable (<500ms cold start)

## Performance Expectations

### Cold Start Times (Approximate)
- **Spain**: ~200ms (11K codes)
- **France**: ~300ms (20K codes)
- **Germany**: ~200ms (11K codes)
- **Portugal**: ~1000ms (197K codes)

### Operation Latencies
- **Geocode by postal**: <50ms (warm)
- **Reverse geocode**: 100-500ms
- **Validate**: <10ms
- **Autocomplete**: <50ms

## Next Steps After Implementation

1. **Monitor performance** in production
2. **Set up alerts** for latency and error rates
3. **Update API documentation** with new country endpoints
4. **Consider scaling** (multiple Lambda functions if needed)
5. **Plan next countries** based on demand

---

**Last Updated**: November 2025
**Based on**: France implementation experience
**Status**: Ready for production use