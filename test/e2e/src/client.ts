/**
 * Geocode ES Lambda Client
 * 
 * Client for invoking the pricofy-geocode-pt Lambda function via AWS SDK.
 */

import { LambdaClient, InvokeCommand } from '@aws-sdk/client-lambda';
import { config } from './config';

export interface GeocodeByPostalRequest {
  operation: 'geocode-by-postal';
  postalCode?: string;
  municipality?: string;
}

export interface ReverseGeocodeRequest {
  operation: 'reverse-geocode';
  lat: number;
  lon: number;
}

export interface ValidatePostalRequest {
  operation: 'validate-postal';
  postalCode: string;
}

export interface ValidateMunicipalityRequest {
  operation: 'validate-municipality';
  municipality: string;
}

export interface AutocompletePostalRequest {
  operation: 'autocomplete-postal';
  prefix: string;
  limit?: number;
}

export interface AutocompleteMunicipalityRequest {
  operation: 'autocomplete-municipality';
  query: string;
  limit?: number;
}

export interface LambdaResponse {
  statusCode: number;
  body: any;
}

/**
 * Client for invoking Geocode ES Lambda function
 */
export class GeocodePTClient {
  private client: LambdaClient;
  private functionName: string;

  constructor() {
    this.client = new LambdaClient({ region: config.awsRegion });
    this.functionName = config.lambdaFunctionName;
  }

  /**
   * Invoke Lambda function with request payload
   */
  private async invoke(request: any): Promise<LambdaResponse> {
    const command = new InvokeCommand({
      FunctionName: this.functionName,
      Payload: JSON.stringify({ body: JSON.stringify(request) }),
    });

    const response = await this.client.send(command);
    const payload = JSON.parse(new TextDecoder().decode(response.Payload));

    // Some error responses may not include a body (e.g., 404). Guard against undefined to avoid JSON.parse crashes.
    const body = typeof payload.body === 'string' ? JSON.parse(payload.body) : {};

    return {
      statusCode: payload.statusCode,
      body,
    };
  }

  /**
   * Geocode by postal code or municipality
   */
  async geocodeByPostal(postalCode?: string, municipality?: string): Promise<LambdaResponse> {
    return this.invoke({
      operation: 'geocode-by-postal',
      ...(postalCode && { postalCode }),
      ...(municipality && { municipality }),
    });
  }

  /**
   * Reverse geocode coordinates to postal code
   */
  async reverseGeocode(lat: number, lon: number): Promise<LambdaResponse> {
    return this.invoke({
      operation: 'reverse-geocode',
      lat,
      lon,
    });
  }

  /**
   * Validate postal code
   */
  async validatePostal(postalCode: string): Promise<LambdaResponse> {
    return this.invoke({
      operation: 'validate-postal',
      postalCode,
    });
  }

  /**
   * Validate municipality
   */
  async validateMunicipality(municipality: string): Promise<LambdaResponse> {
    return this.invoke({
      operation: 'validate-municipality',
      municipality,
    });
  }

  /**
   * Autocomplete postal code
   */
  async autocompletePostal(prefix: string, limit?: number): Promise<LambdaResponse> {
    return this.invoke({
      operation: 'autocomplete-postal',
      prefix,
      ...(limit && { limit }),
    });
  }

  /**
   * Autocomplete municipality
   */
  async autocompleteMunicipality(query: string, limit?: number): Promise<LambdaResponse> {
    return this.invoke({
      operation: 'autocomplete-municipality',
      query,
      ...(limit && { limit }),
    });
  }
}

