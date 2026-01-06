#!/usr/bin/env node

/**
 * Wrapper script that cleans the OpenAPI spec and generates mocks
 * This replaces the direct msw-auto-mock call to handle circular references
 */

const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

const main = async () => {
  try {
    console.log('Step 1: Cleaning OpenAPI spec to remove circular references...');
    
    // Run the cleaning script
    const cleanResult = execSync('node ./scripts/clean-openapi-spec.js', {
      encoding: 'utf8',
      cwd: process.cwd(),
      env: { ...process.env, NUON_API_URL: process.env.NUON_API_URL || 'https://api.nuon.co' }
    });
    
    // Extract the cleaned spec file path
    const cleanedSpecFile = path.join(__dirname, 'cleaned-openapi-spec.json');
    
    if (!fs.existsSync(cleanedSpecFile)) {
      throw new Error('Cleaned spec file was not created');
    }
    
    console.log('Step 2: Generating mocks from cleaned spec...');
    
    // Get environment variables
    const baseUrl = process.env.NUON_API_URL || 'https://api.nuon.co';
    console.log(`Using API URL: ${baseUrl}`);
    
    // Generate mocks using the cleaned spec
    const mockCommand = `node --max-old-space-size=8192 ./node_modules/.bin/msw-auto-mock "${cleanedSpecFile}" -o ./test/mock-api-handlers.js -m 3 --node --base-url "${baseUrl}"`;
    
    console.log(`Running: ${mockCommand}`);
    
    const mockResult = execSync(mockCommand, {
      encoding: 'utf8',
      cwd: process.cwd(),
      stdio: 'inherit'
    });
    
    console.log('Step 3: Cleaning up temporary files...');
    
    // Remove the temporary cleaned spec file
    fs.unlinkSync(cleanedSpecFile);
    
    console.log('✅ Mock generation completed successfully!');
    
  } catch (error) {
    console.error('❌ Mock generation failed:', error.message);
    
    // Cleanup on failure
    const cleanedSpecFile = path.join(__dirname, 'cleaned-openapi-spec.json');
    if (fs.existsSync(cleanedSpecFile)) {
      fs.unlinkSync(cleanedSpecFile);
    }
    
    process.exit(1);
  }
};

main();
