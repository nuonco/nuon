# Dashboard UI E2E Tests

End-to-end tests using Playwright for browser-based testing against the real API.

## Prerequisites

1. **Install dependencies**: `npm install`
2. **Install browsers**: `npx playwright install chromium`
3. **Running services**:
   - Dashboard: Running on port 4000
   - API: Running on port 8081
4. **Authentication token**: `E2E_AUTH_TOKEN` environment variable set (created by nuonctl script)

## Running Tests

```bash
# Run all E2E tests
npm run test:e2e

# Interactive UI mode (best for debugging)
npm run test:e2e:ui

# Watch browser execution
npm run test:e2e:headed

# Step-through debugger
npm run test:e2e:debug

# Generate tests from browser actions
npm run test:e2e:codegen
```

## Test Organization

- `/e2e/tests/` - Test spec files (`.spec.ts`)
- `/e2e/fixtures/` - Reusable test fixtures and utilities (future)

## Writing Tests

Tests are written using Playwright's test API. Example:

```typescript
import { test, expect } from '@playwright/test';

test('should do something', async ({ page }) => {
  await page.goto('/');
  await expect(page.locator('h1')).toHaveText('Welcome');
});
```

## Debugging Tips

1. **Use UI mode**: `npm run test:e2e:ui` - Interactive debugging interface
2. **Run headed**: `npm run test:e2e:headed` - See browser window
3. **Use debug mode**: `npm run test:e2e:debug` - Step through tests
4. **Screenshots**: Automatically captured on failure
5. **Videos**: Recorded when tests fail

## Authentication

Tests use the Nuon auth service with token from `E2E_AUTH_TOKEN` environment variable. The token is set as an `X-Nuon-Auth` cookie for authenticated requests.

To run authenticated tests, ensure `E2E_AUTH_TOKEN` is set in your environment before running the tests.
