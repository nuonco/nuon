#!/usr/bin/env node

/**
 * Generate TypeScript types from OpenAPI spec.
 * Supports both remote API endpoints and local spec files.
 *
 * When NUON_API_URL points to localhost, waits for the API to become
 * available before generating (useful when ctl-api starts in parallel).
 *
 * Environment variables:
 * - NUON_OPENAPI_SPEC_FILE: Path to local OpenAPI v3 spec file (takes precedence)
 * - NUON_API_URL: Remote API URL to fetch spec from (default: https://api.nuon.co)
 */

const { execSync } = require('child_process');
const fs = require('fs');

const SPEC_FILE = process.env.NUON_OPENAPI_SPEC_FILE;
const API_URL = process.env.NUON_API_URL || 'https://api.nuon.co';
const OUTPUT_FILE = './client/types/nuon-oapi-v3.d.ts';

const isLocalhost = (url) => {
  try {
    const { hostname } = new URL(url);
    return hostname === 'localhost' || hostname === '127.0.0.1';
  } catch {
    return false;
  }
};

const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

const waitForAPI = async (url, { maxWaitMs = 60_000, intervalMs = 2_000 } = {}) => {
  const start = Date.now();
  console.log(`⏳ Waiting for API at ${url} ...`);
  while (Date.now() - start < maxWaitMs) {
    try {
      const res = await fetch(url);
      if (res.ok) {
        console.log(`✅ API is up`);
        return;
      }
    } catch {}
    await sleep(intervalMs);
  }
  throw new Error(`API at ${url} did not become available within ${maxWaitMs / 1000}s`);
};

const generate = (source) => {
  execSync(`npx openapi-typescript "${source}" -o ${OUTPUT_FILE}`, {
    stdio: 'inherit',
    cwd: process.cwd(),
  });
  console.log(`✅ Generated types at ${OUTPUT_FILE}`);
};

const main = async () => {
  if (SPEC_FILE) {
    if (!fs.existsSync(SPEC_FILE)) {
      console.error(`Error: Spec file not found: ${SPEC_FILE}`);
      process.exit(1);
    }
    console.log(`Generating API types from local file: ${SPEC_FILE}`);
    generate(SPEC_FILE);
    return;
  }

  const source = `${API_URL}/oapi/v3`;

  if (isLocalhost(API_URL)) {
    try {
      await waitForAPI(source);
    } catch (error) {
      if (fs.existsSync(OUTPUT_FILE)) {
        console.warn(`⚠️  ${error.message}, using existing ${OUTPUT_FILE}`);
        return;
      }
      console.error(`❌ ${error.message} and no existing types file found`);
      process.exit(1);
    }
  }

  console.log(`Generating API types from: ${source}`);
  try {
    generate(source);
  } catch (error) {
    if (fs.existsSync(OUTPUT_FILE)) {
      console.warn(`⚠️  Failed to generate API types (${error.message}), using existing ${OUTPUT_FILE}`);
    } else {
      console.error('❌ Failed to generate API types and no existing file found:', error.message);
      process.exit(1);
    }
  }
};

main();
