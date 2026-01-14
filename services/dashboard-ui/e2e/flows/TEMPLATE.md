# [Flow Name] Flow

**Priority**: [High | Medium | Low]
**Status**: ❌ Not Automated
**Test File**: `tests/[flow-name].spec.ts`

## Prerequisites

Before starting this flow, ensure:
- [ ] User is authenticated
- [ ] [What data needs to exist - e.g., "User has at least one app configured"]
- [ ] [What page/state user starts from - e.g., "User is on the dashboard home page"]

## User Flow

### Step 1: [Action Description]

*(Optional screenshot: `screenshots/[flow-name]-1.png`)*

- **User Action**: [What the user does - e.g., "Click the 'Create Install' button"]
- **Expected Result**: [What should happen - e.g., "Modal opens with install creation form"]
- **Key Selectors**:
  - Button/Link: `[visible text or data-testid]`
  - Form field: `[name or id attribute]`

### Step 2: [Next Action]

*(Optional screenshot: `screenshots/[flow-name]-2.png`)*

- **User Action**: [What happens next]
- **Expected Result**: [What should happen]
- **Key Selectors**:
  - Element: `[selector]`

### Step 3: [Continue...]

*(Repeat for each step in the flow)*

---

## Test Data

Example data needed to run this flow:

```typescript
const testData = {
  fieldName: 'example-value',
  // Add any test data here
}
```

## Edge Cases to Test

Document scenarios that should be tested beyond the happy path:

- [ ] Empty required fields show validation errors
- [ ] Submitting invalid data shows error message
- [ ] Canceling the flow closes modal/returns to previous page
- [ ] [Add more edge cases specific to this flow]

## Notes

Any additional context, gotchas, or important information:

- [Note about special behavior]
- [Dependencies on other flows or features]
- [Known issues or limitations]

---

## For Developers

**Selector Recommendations**:
- Buttons: `page.getByRole('button', { name: 'Button Text' })`
- Form fields: `page.locator('[name="field-name"]')` or `page.locator('#field-id')`
- Headings: `page.getByRole('heading', { name: 'Heading Text' })`
- Links: `page.getByRole('link', { name: 'Link Text' })`
- Text content: `page.getByText('Exact text')`

**Test Structure Example**:

```typescript
import { test, expect } from '../fixtures/auth.fixture';

test.describe('[Flow Name]', () => {
  test('should complete happy path', async ({ authenticatedPage }) => {
    // Step 1: [Action]
    await authenticatedPage.goto('/start-page');

    // Step 2: [Action]
    await authenticatedPage.getByRole('button', { name: 'Button' }).click();

    // Step 3: Verify result
    await expect(authenticatedPage).toHaveURL(/expected-url/);
  });

  test('should handle edge case', async ({ authenticatedPage }) => {
    // Test edge case scenario
  });
});
```
