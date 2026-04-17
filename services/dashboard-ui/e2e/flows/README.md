# Flow Spec Format

Structured markdown format for describing E2E test flows. These serve as source-of-truth documentation — when flows change, update the markdown and ask Claude Code to regenerate the Playwright spec.

## Format

```markdown
# Flow: <name>

## Setup
- env: E2E_ORG_ID (required)
- start: /:orgId/installs

## Steps

### <step name>
- action: goto | /:orgId/apps
- expect: visible | heading "Apps"

### <step name>
- action: click | button "Create install"
- expect: visible | text "Select an app"
```

## Action types

| Action | Syntax | Description |
|--------|--------|-------------|
| `goto` | `goto \| /path` | Navigate to URL |
| `click` | `click \| button "Label"` | Click element by role + name |
| `fill` | `fill \| input "Label" \| value` | Fill input by label |
| `select` | `select \| select "Label" \| option` | Select dropdown option |
| `wait` | `wait \| networkidle` | Wait for condition |

## Assertion types

| Assertion | Syntax | Description |
|-----------|--------|-------------|
| `visible` | `visible \| heading "Text"` | Element is visible |
| `not-visible` | `not-visible \| text "Error"` | Element is not visible |
| `title` | `title \| "Page Title"` | Page title matches |
| `url` | `url \| /apps` | URL contains path |
| `count` | `count \| row \| 3` | Element count matches |

## Locator types

Used in actions and assertions after the `|` separator:

- `heading "Text"` — `getByRole('heading', { name: 'Text' })`
- `button "Text"` — `getByRole('button', { name: 'Text' })`
- `link "Text"` — `getByRole('link', { name: 'Text' })`
- `text "Text"` — `getByText('Text')`
- `input "Label"` — `getByLabel('Label')`
- `select "Label"` — `getByLabel('Label')`
- `testid "id"` — `getByTestId('id')`
- `row` — `getByRole('row')`
- `.class-name` — `locator('.class-name')`
