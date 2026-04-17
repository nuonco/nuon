# Dashboard UI Scripts

## Scripts

### `generate-api-types.js`
Generates TypeScript types from the Nuon OpenAPI spec into `client/types/nuon-oapi-v3.d.ts` using `openapi-typescript`. Runs automatically before `dev` and `tsc`.

### `clean-openapi-spec.js`
Downloads the OpenAPI spec and removes circular `$ref` chains that break `msw-auto-mock`. Outputs a cleaned spec to `scripts/cleaned-openapi-spec.json`.

### `hash-assets.js`
Post-build step that content-hashes JS and CSS output files, rewrites `index.html` references to point at the hashed filenames in `dist/assets/`, and cleans up unhashed originals. Runs automatically as part of `npm run build`.

### `dev.sh`
Starts the Go BFF server (`server/`), waits for it to write a port file, then runs the SPA dev process (`npm run dev`) with live reload.

## Usage

### Generate API types (production API, default)
```bash
npm run generate-api-types
```

### Generate API types from local API
```bash
NUON_API_URL=http://localhost:8081 npm run generate-api-types
```

### Generate API types from a local spec file
```bash
NUON_OPENAPI_SPEC_FILE=./path/to/spec.json npm run generate-api-types
```

### Clean OpenAPI spec (for mock generation)
```bash
node scripts/clean-openapi-spec.js
# or with a local API:
NUON_API_URL=http://localhost:8081 node scripts/clean-openapi-spec.js
# or with a local file:
NUON_OPENAPI_SPEC_FILE=./path/to/spec.json node scripts/clean-openapi-spec.js
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `NUON_API_URL` | `https://api.nuon.co` | API URL to fetch the OpenAPI spec from |
| `NUON_OPENAPI_SPEC_FILE` | — | Local spec file path (takes precedence over `NUON_API_URL`) |
