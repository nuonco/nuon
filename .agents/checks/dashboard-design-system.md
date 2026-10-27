---
name: Dashboard-design-system
description: Flags production dashboard UI that invents components or breaks Stratus instead of reusing the existing system
severity-default: high
tools: [Grep, Read]
---

Applies to changes under `services/dashboard-ui/client/views/` and `services/dashboard-ui/client/components/`. Skip `onboarding/`, `playground/`, and `client/lite/`. Onboarding is marketing-owned; visual drift there is not a defect.

Read `services/dashboard-ui/AGENTS.md` (Design system), `DESIGN.md` §0 and §4, and `client/components/PATTERNS.md`.

Look for:

- A new visual pattern, color, page shell, or component when PATTERNS.md already names the primitive and a shipped page to copy.
- Stock Tailwind palettes (`neutral`, `slate`, `gray`, `zinc`, and the rest outside Stratus), raw hex, arbitrary `px` spacing, `border-2` or a colored border, `space-y-*` / `space-x-*`.
- A raw `<button>`, `<input>`, `<select>`, or `<textarea>` in a view or feature component.
- A direct `Skeleton`, `TableSkeleton`, or `TimelineSkeleton` in a feature, or a new `*Skeleton` component.
- `PageHeader`, `HeadingGroup`, or `PageHeadingGroup` assembled in a view. History rendered as a column beside the page.
- An icon import from `@phosphor-icons/react`, `lucide-react`, `@heroicons/*`, or `react-icons` outside `components/common/Icon.tsx`.
- A `Select` or `FormSelect` driving a list/table filter instead of `RadioFilterDropdown` / `CheckboxFilterDropdown`.
- `size="sm"` / `size="xs"` (or any non-default `size`) with no reason stated in the PR.
- `variant="ghost"` on an ordinary action (a submit, a create button, a row action) rather than UI chrome (sidebar toggle, diff/row expand, `ShowMore`).
- A hand-rolled segmented control (`aria-pressed` buttons) instead of `ToggleButton` / `Tabs` / `TabNav`.
- A hand-rolled content clip (a `max-height` / `mask-image` fade with its own toggle button) instead of wrapping the content in `ShowMore` from `common/ShowMore`. The small label + icon on a full-row collapsible header (DESIGN.md "Collapsible header") is not this — that one stays inline `Text` + `Icon` and is not a defect.
- A raised count or a new path in `client/guardrails/design-system.baseline.json`. That file only shrinks. If the file is not in the change, skip this bullet.

Report the file, the rule it breaks, and the component or shipped page to use instead.
