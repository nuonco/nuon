---
name: dashboard-ui:announcement
description: Add a new announcement to the dashboard-ui org dashboard page
---

You are adding a new announcement entry to `services/dashboard-ui/client/content/dashboard-announcements.json`. This file powers the announcement cards on the org dashboard page.

## Step 1: Gather context from git

Before asking the user anything, run these commands to gather smart defaults:

1. `git describe --tags --abbrev=0` — Get the last git tag name (use as option 1 for title)
2. `git log -1 --pretty=%s` — Get the latest commit message (use as option 2 for title)

Store both values — you'll use them as suggested options in the questions below.

## Step 2: Gather details from user

Use **AskUserQuestion** to ask the user. The tool requires 2-4 options per question, so use the git context from Step 1 to generate smart suggestions. The user can always pick "Other" for free text.

Ask in batches of up to 4 questions (the tool's max). Run multiple rounds if needed.

### Round 1:

1. **Title** (header: "Title")
   - Option 1: Title derived from the last git tag (clean it up — strip `v` prefix, convert hyphens/underscores to spaces, sentence case)
   - Option 2: Title derived from the latest commit message (clean it up to sentence case)
   - The user can pick "Other" to type their own

2. **Date** (header: "Date")
   - Option 1: Today's date formatted as `MMM DD, YYYY` (label it with the actual date, e.g. "Today (Apr 29, 2026)")
   - Option 2: "Custom date" — description says "Select Other to type a specific date in MMM DD, YYYY format"

3. **Description** (header: "Description")
   - Option 1: Generate a 1-sentence description based on the git tag or recent commits
   - Option 2: Generate an alternative 1-sentence description with a different angle
   - The user can pick "Other" to type their own

4. **CTA text** (header: "CTA text")
   - Option 1: "Learn more" (mark as recommended)
   - Option 2: "Try now"

### Round 2:

5. **CTA URL** (header: "CTA URL")
   - Scan `docs/updates/` for `.mdx` files and find the one with the highest numeric prefix (e.g. `028-open-source.mdx`). Strip the `.mdx` extension to get the slug.
   - Option 1: `https://docs.nuon.co/updates/<highest-slug>` — description: "Latest changelog entry found in docs/updates/". For example if the highest file is `028-open-source.mdx`, the URL would be `https://docs.nuon.co/updates/028-open-source`
   - Option 2: `https://docs.nuon.co/` — description: "Nuon docs root — use Other to type a specific URL"

6. **Image filenames** (header: "Images")

   Before asking, scan the images directory for unused images:
   - List all files in `services/dashboard-ui/public/images/announcements/`
   - Read the current `dashboard-announcements.json` and collect all `image` and `imageDark` paths
   - Find any image files that are NOT referenced in the JSON (these are likely new images the user added for this announcement)

   Options (always exactly 2):
   - Option 1: If unreferenced images were found, list them as a pair (e.g. `feature-light.png / feature-dark.png`). Description: "Found unreferenced images in the announcements directory". If no unreferenced images exist, use auto-generated filenames based on the slugified title instead (e.g. `<slug>-light.png / <slug>-dark.png`). Description: "Auto-generated from title. Place files in services/dashboard-ui/public/images/announcements/"
   - Option 2: Auto-generated filenames based on the slugified title (e.g. `<slug>-light.png / <slug>-dark.png`). Description: "Auto-generated from title. Place files in services/dashboard-ui/public/images/announcements/". If Option 1 already uses auto-generated names (no unreferenced images found), instead use a variation like `<slug>-screenshot-light.png / <slug>-screenshot-dark.png`.

When the user picks an image option, use the selected or provided filenames for both light and dark mode images.

Wait for all user responses before proceeding to Step 3.

## Step 3: Generate the entry

Build the announcement object with these fields:

- **`id`** — Generate from a slugified version of the title + a 3-digit sequence number. Look at the last entry's id to determine the next sequence number. Example pattern: `nuon-tuis-extensions-032`
- **`title`** — From user input
- **`date`** — From user input (or today's date)
- **`description`** — From user input
- **`image`** — `/images/announcements/<light-mode-filename>`
- **`imageDark`** — `/images/announcements/<dark-mode-filename>`
- **`ctaText`** — From user input (or "Learn more")
- **`ctaUrl`** — From user input
- **`dismissible`** — Always `false`

## Step 4: Add to the JSON file

1. Read `services/dashboard-ui/client/content/dashboard-announcements.json`
2. **Set `dismissible: true` on all existing entries** — only the newest announcement should be non-dismissible
3. **Prepend** the new entry (with `dismissible: false`) to the beginning of the `announcements` array (newest first)
4. Write the updated JSON file

## Validation

- The `ctaUrl` must start with `https://`
- Image filenames should have an extension (`.png`, `.jpg`, `.svg`, etc.)
- Title and description should use sentence case (not Title Case)

## Anti-patterns

- **Never leave old entries as `dismissible: false`** — only the newest (first) entry should be non-dismissible; all others must be `dismissible: true`
- **Never append to the end of the array** — newest announcements go first
- **Never hardcode the image directory path differently** — always use `/images/announcements/`
- **Never skip asking about both light and dark mode images** — both are needed for proper theme support
