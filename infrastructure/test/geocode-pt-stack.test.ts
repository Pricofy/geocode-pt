/**
 * CDK tests for Geocode PT Stack
 */

import * as cdk from 'aws-cdk-lib';
import { Template, Match } from 'aws-cdk-lib/assertions';
import { GeocodePtStack } from '../lib/geocode-pt-stack';

describe('Geocode PT Stack', () => {
  let app: cdk.App;
  let stack: GeocodePtStack;
  let template: Template;

  beforeEach(() => {
    // Create dummy dist directory for tests (Go binary)
    const fs = require('fs');
    const path = require('path');
    const distDir = path.join(__dirname, '../../dist');
    fs.mkdirSync(distDir, { recursive: true });
    
    // Only create dummy bootstrap if it doesn't exist (don't overwrite real binary)
    const bootstrapPath = path.join(distDir, 'bootstrap');
    if (!fs.existsSync(bootstrapPath)) {
      // Create dummy bootstrap binary (Go Lambda entry point)
      fs.writeFileSync(bootstrapPath, '#!/bin/sh\necho "dummy"');
    }

    app = new cdk.App();
    stack = new GeocodePtStack(app, 'TestStack', {
      env: { account: 'test-account', region: 'eu-west-1' },
      environment: 'dev',
    });
    template = Template.fromStack(stack);
  });

  it('should create exactly one Lambda function', () => {
    // Single geocode Lambda with internal routing
    template.resourceCountIs('AWS::Lambda::Function', 1);
  });

  it('should create Lambda with proper naming', () => {
    template.hasResourceProperties('AWS::Lambda::Function', {
      FunctionName: 'pricofy-geocode-pt',
    });
  });

  it('should configure Lambda with appropriate memory', () => {
    // 256MB required for Portugal (197K postal codes, 23MB JSON)
    template.hasResourceProperties('AWS::Lambda::Function', {
      MemorySize: 256,
    });
  });

  it('should configure Lambda with appropriate timeout', () => {
    // 10 seconds should be enough for all operations
    template.hasResourceProperties('AWS::Lambda::Function', {
      Timeout: 10,
    });
  });

  it('should configure Lambda with proper runtime', () => {
    template.hasResourceProperties('AWS::Lambda::Function', {
      Runtime: 'provided.al2023',
    });
  });

  it('should configure Lambda with handler', () => {
    template.hasResourceProperties('AWS::Lambda::Function', {
      Handler: 'bootstrap',
    });
  });

  it('should configure environment variables', () => {
    template.hasResourceProperties('AWS::Lambda::Function', {
      Environment: {
        Variables: {
          ENVIRONMENT: 'dev',
        },
      },
    });
  });

  it('should enable X-Ray tracing', () => {
    template.hasResourceProperties('AWS::Lambda::Function', {
      TracingConfig: {
        Mode: 'Active',
      },
    });
  });
});

