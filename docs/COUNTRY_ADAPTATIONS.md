# Country-Specific Adaptations

This document outlines the specific adaptations required for each country when extending the geocode service from Spain.

## Overview

Each country requires specific adaptations in:

1. **Postal Code Validation**: Regex patterns for format validation
2. **Administrative Mapping**: How to map GeoNames admin fields to "province"
3. **Postal Code Normalization**: Special handling for non-standard formats
4. **Lambda Configuration**: Memory requirements and performance considerations

## Country Adaptations

### España (ES) - Base Implementation

**Postal Code Format:**
- Regex: `^\d{5}$`
- Example: `28001`
- Validation: 5 digits only

**Administrative Mapping:**
- Province: `admin name2` (Province)
- Municipality: `place name`

**Lambda Configuration:**
- Memory: 128MB
- Dataset: ~11,150 codes (~1.5MB JSON)

### Francia (FR)

**Postal Code Format:**
- Regex: `^\d{5}(\s+[A-Z\s]+)?$`
- Example: `75001` or `75021 CEDEX 01`
- Validation: 5 digits, optionally followed by space and uppercase text (CEDEX, etc.)
- Normalization: Remove suffixes, keep only 5-digit code

**Administrative Mapping:**
- Province: `admin name2` (Département)
- Municipality: `place name`
- Examples:
  - Paris → Paris (département)
  - Marseille → Bouches-du-Rhône

**Lambda Configuration:**
- Memory: 128MB
- Dataset: ~20,414 codes (~2.4MB JSON)
- Cold Start: ~300-400ms

### Alemania (DE)

**Postal Code Format:**
- Regex: `^\d{5}$`
- Example: `10115`
- Validation: 5 digits only

**Administrative Mapping:**
- Province: `admin name1` (State/Land)
- Municipality: `place name`
- Examples:
  - Berlin → Land Berlin
  - Munich → Bayern (Bavaria)

**Lambda Configuration:**
- Memory: 128MB
- Dataset: ~10,813 codes (~1.3MB JSON)
- Cold Start: ~200ms (similar to Spain)

### Italia (IT)

**Postal Code Format:**
- Regex: `^\d{5}$`
- Example: `00118`
- Validation: 5 digits only

**Administrative Mapping:**
- Province: `admin name2` (Provincia)
- Municipality: `place name`
- Examples:
  - Rome → Roma (province)
  - Milan → Milano

**Lambda Configuration:**
- Memory: 128MB
- Dataset: ~4,484 codes (~1MB JSON)
- Cold Start: ~150-200ms

### Portugal (PT)

**Postal Code Format:**
- Raw Format: `XXXX-XXX` (7 digits with hyphen)
- Regex Validation: `^\d{4}-\d{3}$|^\d{7}$` (accept both formats)
- Example: `1000-205` or `1000205`
- Normalization: Remove hyphen, store as `XXXXXXX`

**Administrative Mapping:**
- Province: `admin name1` (District)
- Municipality: `place name`
- Examples:
  - Lisbon → Lisboa (district)
  - Porto → Porto

**Lambda Configuration:**
- Memory: 256MB ⚠️
- Dataset: ~169,712 codes (~25MB JSON)
- Cold Start: ~500-700ms
- Special consideration: Larger dataset requires more memory

## Implementation Checklist

### 1. Data Conversion Script Updates

The `scripts/convert-geonames-to-json.js` script handles:

- [x] **Portugal normalization**: Remove hyphens from postal codes
- [x] **France normalization**: Remove CEDEX suffixes
- [ ] **Future**: Add country-specific normalization as needed

### 2. Validation Script Updates

The `scripts/validate-postal-data.js` script includes:

- [x] **Country-specific regex patterns** for postal code validation
- [x] **Coordinate range validation** (-90 to 90 lat, -180 to 180 lon)
- [x] **Administrative field validation** (municipality, province)
- [ ] **Future**: Add specific business rules per country

### 3. Go Code Adaptations

For each country service, update:

#### `internal/domain/constants.go`
```go
const (
    // Postal code regex patterns by country
    PostalCodeRegexES = `^\d{5}$`
    PostalCodeRegexFR = `^\d{5}(\s+[A-Z\s]+)?$`
    PostalCodeRegexDE = `^\d{5}$`
    PostalCodeRegexIT = `^\d{5}$`
    PostalCodeRegexPT = `^\d{4}-\d{3}$|^\d{7}$`
)
```

#### `internal/infrastructure/provider/postal_provider.go`
- Replace `postal-codes-es.json` with country-specific file
- Update embed directive: `//go:embed postal-codes-{country}.json`

#### `internal/domain/models.go`
- Update comments to reflect country-specific administrative divisions
- Example: "Province (Département for France, State for Germany, etc.)"

### 4. Infrastructure (CDK) Adaptations

#### `infrastructure/lib/geocode-{country}-stack.ts`
- Update class name: `GeocodeEsStack` → `Geocode{Country}Stack`
- Update function name: `pricofy-geocode-es` → `pricofy-geocode-{country}`
- Update exports: `Pricofy-GeocodeEsArn` → `Pricofy-Geocode{Country}Arn`
- Update descriptions and comments
- Adjust memory size for Portugal (256MB)

#### `infrastructure/test/geocode-{country}-stack.test.ts`
- Update stack name references
- Update function name references
- Update export references

### 5. CI/CD Adaptations

#### GitHub Actions Workflows
- Update workflow names and descriptions
- Update Lambda function names in E2E tests
- Update deployment references

#### Tests E2E
- Update `test/e2e/src/config.ts` with correct function name
- Update `test/e2e/src/geocode-{country}.test.ts` with:
  - Country-specific postal code examples
  - Correct municipality/province expectations
  - Valid coordinate ranges for the country

### 6. Documentation Updates

#### `README.md`
- Update service description for specific country
- Update API examples with country-specific postal codes
- Update coverage information

#### `docs/POSTAL_CODES.md`
- Update statistics (total codes, size, ranges)
- Update data source information
- Document country-specific format quirks

#### `docs/CONTRACTS.md`
- Update API examples with real postal codes
- Update response examples with country-specific data

## Performance Considerations

| Country | Codes | JSON Size | Memory | Cold Start | Notes |
|---------|-------|-----------|--------|------------|-------|
| España | 11,150 | 1.5MB | 128MB | 200ms | Baseline |
| Francia | 20,414 | 2.4MB | 128MB | 300-400ms | CEDEX handling |
| Alemania | 10,813 | 1.3MB | 128MB | 200ms | Similar to ES |
| Italia | 4,484 | 1MB | 128MB | 150-200ms | Smallest dataset |
| Portugal | 169,712 | 25MB | 256MB | 500-700ms | Largest dataset |

## Migration Strategy

1. **Phase 1**: Clone Spain service → France (similar structure)
2. **Phase 2**: Clone France → Germany (refine process)
3. **Phase 3**: Clone Germany → Italy (smaller dataset)
4. **Phase 4**: Clone Italy → Portugal (handle large dataset + special format)

## Quality Assurance

### Automated Checks
- [ ] Postal code format validation per country
- [ ] Coordinate range validation
- [ ] Administrative field completeness
- [ ] Memory usage testing
- [ ] Cold start performance testing

### Manual Validation
- [ ] Sample postal codes work correctly
- [ ] Edge cases handled (CEDEX, special formats)
- [ ] Coordinate accuracy verification
- [ ] Administrative division correctness

## Future Considerations

### Potential Improvements
1. **Unified Service**: Multi-country Lambda with routing
2. **Caching**: Redis for frequently accessed codes
3. **Batch Operations**: Process multiple codes at once
4. **Address Parsing**: Parse full addresses to postal codes
5. **Reverse Geocoding**: Improve accuracy with country-specific logic

### Scalability
- **Horizontal Scaling**: Separate Lambda per country allows independent scaling
- **Cost Optimization**: Pay only for traffic per country
- **Maintenance**: Independent deployments reduce risk

---

**Last Updated:** November 2025
**Status:** Ready for implementation