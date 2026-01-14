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
- `/e2e/flows/` - User flow documentation (markdown files)
- `/e2e/fixtures/` - Reusable test fixtures (auth helpers, utilities)

## User Flows

User flows are documented in the `/e2e/flows/` directory. These documents describe critical user journeys and serve as the source of truth for E2E tests.

**Flow Documentation**:
- See `/e2e/flows/README.md` for catalog of all flows
- Each flow corresponds to a test file in `/e2e/tests/`
- Flows can be written by non-technical team members using the template

**Writing User Flows**:
1. Copy `/e2e/flows/TEMPLATE.md` to create a new flow
2. Fill in the prerequisites, steps, test data, and edge cases
3. Submit PR with the markdown file
4. Developer writes corresponding test in `/e2e/tests/`

**Using the Authentication Fixture**:
```typescript
import { test, expect } from '../fixtures/auth.fixture';

test('my test', async ({ authenticatedPage }) => {
  await authenticatedPage.goto('/installs');
  // Already authenticated - no need to set cookies manually!
});
```

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
