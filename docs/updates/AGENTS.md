# Changelog/Updates Directory

This directory contains product update/changelog entries for the Nuon documentation.

## Adding a New Update Entry

To add a new changelog entry, you must update **3 files**:

### 1. Create the MDX file

Create a new file with the next sequential number: `XXX-update-name.mdx`

Use this template:

```mdx
---
title: XXX - Update Title
description: Brief description of the update.
---

_Month DD, YYYY_

<div className="badge badge--primary">vX.XX.XXX</div>

## Main Section

Content here...

## Bug Fixes

- Fix 1
- Fix 2
```

### 2. Update `updates.mdx`

Add a new `<Card>` entry at the **top** of the `<CardGroup>` (after line 10):

```jsx
<Card title="XXX - Update Title" icon="icon-name" href="/updates/XXX-update-name">
  Brief description.
</Card>
```

Common icons used: `gear`, `clock`, `code`, `check`, `lock`, `cloud`, `user`, `layer-group`, `diagram-project`, `trash`

### 3. Update `docs.json`

Add the new page to the Changelog tab's `pages` array, right after `"updates/updates"`:

```json
"updates/XXX-update-name",
```

This is located in `/docs/docs.json` under `navigation.tabs[1].pages` (the Changelog tab).

### 4. Images (if any)

Place images in the `assets/` subdirectory using one of these patterns:
- `assets/XXX_image_name.png` - flat naming
- `assets/XXX/image_name.png` - subfolder per update

Reference in MDX as:
```mdx
![Alt text](assets/XXX_image_name.png)
```
