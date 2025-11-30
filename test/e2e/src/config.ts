/**
 * E2E Test Configuration
 * 
 * Manages environment variables and configuration for end-to-end tests.
 * All values have sensible defaults for CI/CD environments.
 */

/**
 * Test configuration interface
 */
export interface TestConfig {
  /** AWS region where Lambda is deployed */
  awsRegion: string;
  /** Lambda function name to test (always uses name, never ARN or ID) */
  lambdaFunctionName: string;
  /** Environment (dev/prod) */
  environment: string;
  /** Test timeout in milliseconds */
  testTimeout: number;
}

/**
 * Get test configuration from environment variables
 * 
 * The Lambda function name is 'pricofy-geocode-es' for both dev and prod environments.
 * Environment is differentiated by AWS account (via AWS_PROFILE), not by function name.
 * 
 * We search by function name, not by ARN or ID, to ensure tests work after redeployments.
 * 
 * AWS credentials are determined by:
 * - AWS_PROFILE environment variable (e.g., 'pricofy-dev' or 'pricofy-prod')
 * - Or AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY if provided
 * - Or default AWS credential chain
 * 
 * @returns Test configuration object
 */
export function getConfig(): TestConfig {
  const environment = process.env.ENVIRONMENT || process.env.ENV || 'dev';
  
  // Validate AWS_PROFILE matches environment (if both are set)
  if (process.env.AWS_PROFILE && environment) {
    const profile = process.env.AWS_PROFILE;
    if (environment === 'dev' && !profile.includes('dev')) {
      console.warn(`⚠️  Warning: ENVIRONMENT=${environment} but AWS_PROFILE=${profile} (expected pricofy-dev)`);
    }
    if (environment === 'prod' && !profile.includes('prod')) {
      console.warn(`⚠️  Warning: ENVIRONMENT=${environment} but AWS_PROFILE=${profile} (expected pricofy-prod)`);
    }
  }
  
  // Lambda function name is the same across environments (environment is differentiated by AWS account)
  // The function is deployed as 'pricofy-geocode-pt' in both dev and prod accounts
  const defaultFunctionName = 'pricofy-geocode-pt';
  
  return {
    awsRegion: process.env.AWS_REGION || 'eu-west-1',
    // Lambda function name: use explicit name or default
    // We use the function name, not ARN or ID, so tests work after redeployments
    // Note: Environment is differentiated by AWS account, not by function name
    lambdaFunctionName: process.env.LAMBDA_FUNCTION_NAME || defaultFunctionName,
    environment,
    testTimeout: parseInt(process.env.TEST_TIMEOUT || '60000', 10),
  };
}

/**
 * Validate test configuration
 * 
 * Ensures all required configuration is present and valid.
 * Throws error if configuration is invalid.
 */
export function validateConfig(): void {
  const config = getConfig();

  if (!config.awsRegion) {
    throw new Error('AWS_REGION is required');
  }

  if (!config.lambdaFunctionName) {
    throw new Error('LAMBDA_FUNCTION_NAME is required');
  }

  if (!['dev', 'prod'].includes(config.environment)) {
    throw new Error(`ENVIRONMENT must be 'dev' or 'prod', got: ${config.environment}`);
  }

  if (config.testTimeout < 1000) {
    throw new Error('TEST_TIMEOUT must be at least 1000ms');
  }
}

/**
 * Default test configuration
 */
export const config = getConfig();

