#!/usr/bin/env node

/**
 * Generate TypeScript types from OpenAPI spec.
 * Supports both remote API endpoints and local spec files.
 * 
 * Environment variables:
 * - NUON_OPENAPI_SPEC_FILE: Path to local OpenAPI v3 spec file (takes precedence)
 * - NUON_API_URL: Remote API URL to fetch spec from (default: https://api.nuon.co)
 */

const { execSync } = require('child_process');
const fs = require('fs');

const SPEC_FILE = process.env.NUON_OPENAPI_SPEC_FILE;
const API_URL = process.env.NUON_API_URL || 'https://api.nuon.co';
const OUTPUT_FILE = './src/types/nuon-oapi-v3.d.ts';

const main = () => {
  let source;

  if (SPEC_FILE) {
    if (!fs.existsSync(SPEC_FILE)) {
      console.error(`Error: Spec file not found: ${SPEC_FILE}`);
      process.exit(1);
    }
    source = SPEC_FILE;
    console.log(`Generating API types from local file: ${SPEC_FILE}`);
  } else {
    source = `${API_URL}/oapi/v3`;
    console.log(`Generating API types from remote: ${source}`);
  }

  try {
    execSync(`npx openapi-typescript "${source}" -o ${OUTPUT_FILE}`, {
      stdio: 'inherit',
      cwd: process.cwd(),
    });
    console.log(`✅ Generated types at ${OUTPUT_FILE}`);
  } catch (error) {
    console.error('❌ Failed to generate API types:', error.message);
    process.exit(1);
  }
};

main();
