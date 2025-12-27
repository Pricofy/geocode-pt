import * as cdk from 'aws-cdk-lib';
import * as events from 'aws-cdk-lib/aws-events';
import * as targets from 'aws-cdk-lib/aws-events-targets';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as logs from 'aws-cdk-lib/aws-logs';
import * as path from 'path';
import { Construct } from 'constructs';

export interface GeocodePtStackProps extends cdk.StackProps {
  environment: 'dev' | 'prod';
}

/**
 * Pricofy Geocode PT Stack
 *
 * Provides a single Lambda function for all Portuguese postal code operations:
 * - geocode-by-postal
 * - reverse-geocode
 * - validate-postal
 * - validate-municipality
 * - autocomplete-postal
 * - autocomplete-municipality
 *
 * All operations are routed internally by the handler based on the 'operation' parameter.
 *
 * The function is PRIVATE (no Function URLs, no API Gateway).
 * Only invokable by pricofy-location-service via IAM role.
 *
 * Security:
 * - No public endpoints
 * - IAM-based invocation only
 * - Resource-Based Policies configured for pricofy-location-service
 * - Invoked by pricofy-location-service as orchestrator
 *
 * Note: Resource-Based Policies are configured via pricofy-infra stack
 * to avoid circular dependencies. This stack exports function ARN,
 * and pricofy-infra grants permissions to pricofy-location-service role.
 */
export class GeocodePtStack extends cdk.Stack {
  public readonly geocodeFunction: lambda.Function;

  constructor(scope: Construct, id: string, props: GeocodePtStackProps) {
    super(scope, id, props);

    // ===========================================
    // Lambda: geocode (routing handler)
    // ===========================================

    // Create log group explicitly to avoid deprecated logRetention
    const logGroup = new logs.LogGroup(this, 'GeocodeLogGroup', {
      logGroupName: `/aws/lambda/pricofy-geocode-pt`,
      retention: logs.RetentionDays.ONE_MONTH, // GDPR compliance (30 days)
      removalPolicy: cdk.RemovalPolicy.DESTROY,
    });

    this.geocodeFunction = new lambda.Function(this, 'GeocodeFunction', {
      code: lambda.Code.fromAsset(path.join(__dirname, '../../dist')),
      handler: 'bootstrap', // Go Lambda entry point
      runtime: lambda.Runtime.PROVIDED_AL2023, // Go custom runtime (AL2023)
      architecture: lambda.Architecture.ARM_64, // Graviton2: 20% cheaper, better performance
      timeout: cdk.Duration.seconds(10),
      memorySize: 256, // 256MB required for Portugal (24MB JSON database)
      environment: {
        ENVIRONMENT: props.environment,
      },
      description: 'Portuguese postal code geocoding operations (routing handler) - Go implementation',
      functionName: 'pricofy-geocode-pt',
      logGroup: logGroup,
      tracing: lambda.Tracing.ACTIVE, // X-Ray tracing for observability
    });

    // ===========================================
    // Lambda Warmup (Cold Start Prevention)
    // ===========================================

    const warmupRule = new events.Rule(this, 'WarmupRule', {
      ruleName: `pricofy-geocode-pt-warmup-${props.environment}`,
      schedule: events.Schedule.rate(cdk.Duration.minutes(5)),
      description: `Warmup for geocode-pt (${props.environment})`,
    });

    warmupRule.addTarget(
      new targets.LambdaFunction(this.geocodeFunction, {
        event: events.RuleTargetInput.fromObject({
          body: JSON.stringify({
            source: 'warmup',
            concurrency: 2,
          }),
        }),
        retryAttempts: 0,
      })
    );

    // Self-invoke permission for warmup (manual ARN to avoid circular dependency)
    const functionArn = `arn:aws:lambda:${this.region}:${this.account}:function:pricofy-geocode-pt`;
    this.geocodeFunction.addToRolePolicy(
      new iam.PolicyStatement({
        effect: iam.Effect.ALLOW,
        actions: ['lambda:InvokeFunction'],
        resources: [functionArn],
      })
    );

    // ===========================================
    // Exports (for pricofy-location-service to import)
    // ===========================================

    // Export function ARN for pricofy-location-service to configure permissions
    new cdk.CfnOutput(this, 'GeocodePtArn', {
      value: this.geocodeFunction.functionArn,
      exportName: `Pricofy-GeocodePtArn`,
      description: 'ARN of geocode-pt Lambda function',
    });

    new cdk.CfnOutput(this, 'GeocodePtName', {
      value: this.geocodeFunction.functionName,
      exportName: `Pricofy-GeocodePtName`,
      description: 'Name of geocode-pt Lambda function',
    });

    // ===========================================
    // Security Notice
    // ===========================================

    new cdk.CfnOutput(this, 'SecurityNotice', {
      value: 'Resource-Based Policies configured via pricofy-infra stack',
      description: 'Lambda function is private - only invokable by pricofy-location-service IAM role',
    });

    // ===========================================
    // Tags
    // ===========================================

    cdk.Tags.of(this).add('Project', 'Pricofy');
    cdk.Tags.of(this).add('Environment', props.environment);
    cdk.Tags.of(this).add('Component', 'Geocode-PT');
    cdk.Tags.of(this).add('ManagedBy', 'CDK');
  }
}

/**
 * Security Configuration Helper
 * 
 * For pricofy-infra to grant invocation permissions:
 * 
 * ```typescript
 * // In pricofy-infra stack:
 * import { Fn } from 'aws-cdk-lib';
 * 
 * const geocodePtArn = Fn.importValue(`Pricofy-GeocodePtArn`);
 * 
 * // Grant pricofy-location-service role permission to invoke
 * locationServiceRole.addToPolicy(new iam.PolicyStatement({
 *   actions: ['lambda:InvokeFunction'],
 *   resources: [geocodePtArn],
 * }));
 * ```
 * 
 * This approach:
 * - Avoids circular dependencies (geocode-pt doesn't need to know about location-service)
 * - Centralizes access control in pricofy-infra
 * - Follows AWS best practices for cross-stack references
 * - Allows principle of least privilege (only location-service role has access)
 */

