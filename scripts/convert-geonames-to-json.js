#!/usr/bin/env node

/**
 * GeoNames Postal Code Converter
 *
 * Converts GeoNames tab-delimited postal code files to JSON format compatible
 * with pricofy-geocode services.
 *
 * Usage:
 *   node scripts/convert-geonames-to-json.js <input_file> [country_code] > output.json
 *
 * Examples:
 *   node scripts/convert-geonames-to-json.js FR.txt FR > postal-codes-fr.json
 *   node scripts/convert-geonames-to-json.js DE.txt DE > postal-codes-de.json
 *
 * GeoNames format (tab-delimited):
 * 0: country code      : iso country code, 2 characters
 * 1: postal code       : varchar(20)
 * 2: place name        : varchar(180)
 * 3: admin name1       : 1. order subdivision (state) varchar(100)
 * 4: admin code1       : 1. order subdivision (state) varchar(20)
 * 5: admin name2       : 2. order subdivision (county/province) varchar(100)
 * 6: admin code2       : 2. order subdivision (county/province) varchar(20)
 * 7: admin name3       : 3. order subdivision (community) varchar(100)
 * 8: admin code3       : 3. order subdivision (community) varchar(20)
 * 9: latitude          : estimated latitude (wgs84)
 * 10: longitude         : estimated longitude (wgs84)
 * 11: accuracy          : accuracy of lat/lng from 1=estimated to 6=centroid
 *
 * Output JSON format:
 * {
 *   "POSTALCODE": {
 *     "lat": 40.4168,
 *     "lon": -3.7038,
 *     "municipality": "Place Name",
 *     "province": "Admin Name2 or Admin Name1"
 *   }
 * }
 */

const fs = require('fs');
const path = require('path');

function main() {
    const args = process.argv.slice(2);

    if (args.length < 1 || args.length > 2) {
        console.error('Usage: node scripts/convert-geonames-to-json.js <input_file> [country_code]');
        console.error('Examples:');
        console.error('  node scripts/convert-geonames-to-json.js FR.txt FR > postal-codes-fr.json');
        console.error('  node scripts/convert-geonames-to-json.js DE.txt > postal-codes-de.json');
        process.exit(1);
    }

    const inputFile = args[0];
    const countryCode = args[1] || 'XX'; // Default to XX if not specified

    if (!fs.existsSync(inputFile)) {
        console.error(`Error: Input file '${inputFile}' does not exist`);
        process.exit(1);
    }

    console.error(`Converting ${inputFile} for country ${countryCode}...`);

    try {
        const data = fs.readFileSync(inputFile, 'utf8');
        const lines = data.split('\n').filter(line => line.trim().length > 0);

        const postalCodes = {};
        let processedCount = 0;
        let skippedCount = 0;

        for (const line of lines) {
            const fields = line.split('\t');

            // Skip invalid lines
            if (fields.length < 12) {
                console.error(`Warning: Skipping invalid line with ${fields.length} fields: ${line.substring(0, 50)}...`);
                skippedCount++;
                continue;
            }

            const [
                countryCodeField,
                postalCode,
                placeName,
                adminName1,
                adminCode1,
                adminName2,
                adminCode2,
                adminName3,
                adminCode3,
                latitude,
                longitude,
                accuracy
            ] = fields;

            // Skip if not the target country (if specified)
            if (countryCode !== 'XX' && countryCodeField !== countryCode) {
                skippedCount++;
                continue;
            }

            // Validate coordinates
            const lat = parseFloat(latitude);
            const lon = parseFloat(longitude);

            if (isNaN(lat) || isNaN(lon)) {
                console.error(`Warning: Invalid coordinates for postal code ${postalCode}: lat=${latitude}, lon=${longitude}`);
                skippedCount++;
                continue;
            }

            // Validate coordinate ranges
            if (lat < -90 || lat > 90 || lon < -180 || lon > 180) {
                console.error(`Warning: Coordinates out of range for postal code ${postalCode}: lat=${lat}, lon=${lon}`);
                skippedCount++;
                continue;
            }

            // Determine province (prefer admin name2, fallback to admin name1)
            let province = adminName2 || adminName1 || '';

            // Normalize postal code
            let normalizedPostalCode = postalCode;
            if (countryCode === 'PT' && postalCode.includes('-')) {
                // Portugal: remove hyphen
                normalizedPostalCode = postalCode.replace('-', '');
            } else if (countryCode === 'FR' && postalCode.includes(' ')) {
                // France: remove CEDEX and other suffixes
                normalizedPostalCode = postalCode.split(' ')[0];
            }

            // Create entry
            postalCodes[normalizedPostalCode] = {
                lat: parseFloat(lat.toFixed(6)), // Round to 6 decimal places
                lon: parseFloat(lon.toFixed(6)),
                municipality: placeName.trim(),
                province: province.trim()
            };

            processedCount++;
        }

        // Output JSON
        console.log(JSON.stringify(postalCodes, null, 2));

        // Summary
        console.error(`✅ Conversion complete:`);
        console.error(`   Processed: ${processedCount} postal codes`);
        console.error(`   Skipped: ${skippedCount} entries`);
        console.error(`   Total entries in file: ${lines.length}`);
        console.error(`   Output size: ${Object.keys(postalCodes).length} unique postal codes`);

        // Validation summary
        const sampleCodes = Object.keys(postalCodes).slice(0, 5);
        console.error(`   Sample postal codes: ${sampleCodes.join(', ')}`);

    } catch (error) {
        console.error(`Error processing file: ${error.message}`);
        process.exit(1);
    }
}

if (require.main === module) {
    main();
}

module.exports = { main };