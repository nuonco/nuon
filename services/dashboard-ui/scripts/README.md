# Dashboard UI Mock Generation Scripts

These scripts solve the circular reference issue with `msw-auto-mock` when generating mocks from the Nuon OpenAPI specification.

## Scripts

### `clean-openapi-spec.js`
Downloads and cleans the OpenAPI spec by removing circular references that cause `msw-auto-mock` to hang.

### `generate-mocks-with-clean-spec.js`
Orchestrates the full process: cleaning the spec and generating mocks.

## Usage

### Production API (Default)
```bash
npm run generate-api-mocks
```
Uses `https://api.nuon.co` by default.

### Local Development
```bash
NUON_API_URL=http://localhost:8081 npm run generate-api-mocks
```

### Staging Environment
```bash
NUON_API_URL=https://api.nuon-stage.co npm run generate-api-mocks
```

## How It Works

1. **Downloads OpenAPI spec** using HTTP/HTTPS based on the URL protocol
2. **Identifies circular references** like `Account -> Role -> Account`
3. **Replaces problematic schemas** with simplified versions
4. **Flattens `allOf` constructs** that break msw-auto-mock
5. **Generates comprehensive mocks** from the cleaned specification
6. **Cleans up temporary files**

## Output

- Creates `test/mock-api-handlers.js` with hundreds of mock endpoints
- Supports all existing tests without changes
- Preserves the full API structure while breaking infinite loops

## Troubleshooting

### Network Issues
The script will show protocol being used (`http:` or `https:`) and response status codes.

### API Endpoint Issues
Check that your `NUON_API_URL` endpoint serves the OpenAPI spec at `/oapi/v3`.

### Generation Failures
The script includes detailed error messages for parsing, network, and generation issues.