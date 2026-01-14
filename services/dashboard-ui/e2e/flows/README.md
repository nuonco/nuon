# E2E User Flow Documentation

This directory contains documented user flows for the Nuon dashboard. Each flow describes a critical user journey and corresponds to automated E2E tests.

## Flow Status Legend

- ✅ **Automated** - Flow is documented and has corresponding Playwright test
- ⏳ **In Progress** - Flow is documented, test implementation in progress
- ❌ **Not Automated** - Flow is documented but not yet automated

## Critical Flows (Priority: High)

*No flows documented yet. Add your first flow by copying TEMPLATE.md!*

## Secondary Flows (Priority: Medium)

*No flows documented yet.*

## How to Use This Documentation

### For Non-Technical Contributors (Product, Design, QA)

1. **Copy the template**: `cp TEMPLATE.md [your-flow-name].md`
2. **Fill in the flow details**: Describe what the user does step-by-step
3. **Submit for review**: Create a PR with your markdown file
4. **Screenshots optional**: Only add if the UI is complex

### For Developers

1. **Read the flow doc**: `/e2e/flows/[flow-name].md`
2. **Write the test**: `/e2e/tests/[flow-name].spec.ts`
3. **Update this README**: Change status from ❌ to ✅ Automated
4. **Link the test**: Add test file path to the flow entry below

### For Everyone

**Using Claude to write flows**: You can ask Claude to document a flow and write the test:

```
I want to document a user flow for [feature name].

Steps:
1. [User does X]
2. [User does Y]
3. [etc...]

Please create a flow doc using /e2e/flows/TEMPLATE.md
and write a Playwright test in /e2e/tests/.
```

Claude can:
- Read component code to find selectors
- Suggest edge cases
- Generate test code from flow descriptions

## Adding a New Flow

1. Copy `TEMPLATE.md` → `your-flow-name.md`
2. Fill in prerequisites, steps, test data, edge cases
3. Add entry to this README under appropriate priority section
4. Submit PR for review
5. Once approved, developer writes corresponding test
6. Update status from ❌ to ✅ when automated

---

## Upcoming Flows

Flows to consider documenting:

- Create install
- Deploy workflow approval
- Component configuration
- Runner health monitoring
- VCS connection setup
- Build execution
- App configuration sync
