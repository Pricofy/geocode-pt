#!/usr/bin/env node
/**
 * Pricofy Geocode PT - CDK App Entry Point
 *
 * Defines Lambda function for Portuguese postal code geocoding operations.
 */

import 'source-map-support/register';
import * as cdk from 'aws-cdk-lib';
import { GeocodePtStack } from '../lib/geocode-pt-stack';

const app = new cdk.App();

// Get environment from context (defaults to 'dev')
// Support both 'env' and 'environment' for backwards compatibility
const environment = app.node.tryGetContext('env') || app.node.tryGetContext('environment') || 'dev';

// Validate environment
if (!['dev', 'prod'].includes(environment)) {
  throw new Error(`Invalid environment: ${environment}. Must be 'dev' or 'prod'.`);
}

// Common props
const stackProps: cdk.StackProps = {
  env: {
    account: process.env.CDK_DEFAULT_ACCOUNT,
    region: 'eu-west-1',
  },
  tags: {
    Project: "Pricofy",
    Service: "Geocode-PT",
    Environment: environment,
  },
};

// Geocode PT Stack: Lambda function for Portuguese postal code operations
new GeocodePtStack(app, `PricofyGeocodePtStack`, {
  ...stackProps,
  description: `Pricofy Geocode PT (${environment}) - Portuguese postal code geocoding Lambda function`,
  environment,
});

console.log(`✅ Stack name: PricofyGeocodePtStack`);

app.synth();

