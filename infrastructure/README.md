# Pricofy Geocode PT - Infrastructure

This directory contains the AWS CDK infrastructure-as-code for deploying the **pricofy-geocode-pt** service.

## Overview

The infrastructure is defined using AWS CDK (Cloud Development Kit) and deploys:

- **Lambda Function**: Single function (`pricofy-geocode-pt-{env}`) with internal routing
  - Handler: `bootstrap` (Go binary entry point)
  - Runtime: `PROVIDED_AL2023` (Go custom runtime)
  - Memory: 128 MB (optimized for Go, reduced from 256MB Node.js)
  - Timeout: 10 seconds
  - Architecture: x86_64

## Architecture

Unlike `pricofy-location-service`, this service uses a **single Lambda with internal routing**:

```
┌─────────────────────────────────────┐
│  pricofy-geocode-pt Lambda          │
│  (Go binary: bootstrap)             │
│                                     │
│  ┌───────────────────────────────┐ │
│  │  handler.Handler()             │ │
│  │  (Route by operation field)   │ │
│  └───────────────┬───────────────┘ │
│                  │                  │
│         ┌────────┴────────┐         │
│         │                 │         │
│    ┌────▼───┐      ┌─────▼────┐    │
│    │geocode-│      │validate- │    │
│    │by-     │      │postal    │    │
│    │postal  │      └──────────┘    │
│    └────────┘                       │
│         ...more operations...       │
└─────────────────────────────────────┘
```

### Why Single Lambda with Routing?

**Efficiency for Lambda-to-Lambda Invocation:**
- ✅ **Single cold start** instead of 6 separate cold starts
- ✅ **Shared memory** for postal code data (loaded once)
- ✅ **Lower latency** (no need to initialize multiple containers)
- ✅ **Lower cost** (fewer Lambda invocations, less memory allocation)
- ✅ **Simpler deployment** (single artifact)

This service is NOT a REST API - it's invoked directly by `pricofy-location-service` via the AWS Lambda SDK.

## Stack Configuration

### Environment Variables

| Variable       | Description                     | Values         |
|----------------|---------------------------------|----------------|
| `ENVIRONMENT`  | Deployment environment          | `dev`, `prod`  |
| `NODE_ENV`     | Node.js environment             | `production`   |

### Resources

- **Lambda Function**: `pricofy-geocode-pt-{env}`
  - No external layers required (static data bundled)
  - No API Gateway (Lambda-to-Lambda invocation)
  - X-Ray tracing enabled for observability

## Deployment

### Prerequisites

1. AWS CLI configured
2. CDK bootstrapped in target account/region
3. Go 1.21+ installed
4. Dependencies installed:
   ```bash
   # From project root
   make install
   
   # Or manually
   go mod download
   cd infrastructure && npm install
   ```

### Deploy via Make (Recommended)

From the root of `pricofy-geocode-pt`:

```bash
# Development
make deploy ENV=dev

# Production
make deploy ENV=prod
```

### Deploy Directly with CDK

```bash
cd infrastructure
npm install
npm run build
cdk deploy GeocodePtStack-dev --context environment=dev
```

## Testing

Infrastructure tests validate the CDK stack configuration:

```bash
# Run infrastructure tests
npm test

# Run with coverage
npm test -- --coverage
```

Tests verify:
- Lambda function creation
- Proper configuration (memory, timeout, runtime)
- Environment variables
- Tracing configuration

## Stack Outputs

The stack exports:
- Lambda function ARN
- Lambda function name

These are used by `pricofy-location-service` to invoke the geocode operations.

## Cost Optimization

This service is optimized for minimal cost:
- Static data (no database, embedded in binary)
- Small memory footprint (128 MB, 50% reduction from Node.js)
- Fast execution (typically < 100ms)
- Fast cold starts (~200ms vs ~500ms Node.js)
- Single Lambda (shared resources)

## Monitoring

CloudWatch Logs:
```bash
aws logs tail /aws/lambda/pricofy-geocode-pt-dev --follow
```

X-Ray traces:
- Available in AWS Console → X-Ray → Service Map
- Helps debug performance issues

## Troubleshooting

### Lambda Can't Find Data File

**Problem**: This should not occur with Go implementation - data is embedded in binary via `//go:embed`

**Solution**: If issues occur, ensure `make build` compiles correctly:
```bash
make clean
make build
ls -lh dist/bootstrap  # Verify binary exists
```

### Stack Deployment Fails

**Problem**: Missing dependencies or build artifacts

**Solution**: Clean and rebuild:
```bash
make clean
make install
make build
cd infrastructure && npm install && npm run build
cdk deploy GeocodePtStack-dev --context environment=dev
```

**Note**: The Go binary (`dist/bootstrap`) must be built before CDK deployment. CDK packages the binary from `dist/` directory.

## Related Documentation

- [Main README](../README.md) - Service overview
- [CLAUDE.md](../CLAUDE.md) - Detailed technical documentation
- [CHANGELOG.md](../CHANGELOG.md) - Version history
- [API Documentation](../api/openapi.yaml) - Lambda invocation contract

