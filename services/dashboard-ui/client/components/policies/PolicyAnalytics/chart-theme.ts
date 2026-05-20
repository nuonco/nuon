/**
 * Recharts styling shared across the policy analytics charts.
 *
 * Everything here is pinned to the Stratus design system tokens defined in
 * `styles.css` so the charts inherit the same surface colors, border tokens,
 * font family, and typographic scale as the rest of the dashboard:
 *
 *   - Colors      -> `var(--background)`, `var(--foreground)`, `var(--border-color)`
 *                    flip automatically via `prefers-color-scheme: dark`.
 *   - Series      -> `--color-{red,orange,green}-500` (semantic fail / warn / pass).
 *   - Font family -> `--font-inter` (matches `<Text>`).
 *   - Font sizes  -> mirror `<Text>` variants: 11px label / 12px subtext / 14px body.
 *   - Tracking    -> -0.2px (matches body/subtext/label variants).
 *   - Weights     -> 400 normal, 500 strong, 600 stronger (CSS vars).
 */

const FONT_FAMILY =
  'var(--font-inter), -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif'

// `body` variant: 14px / 24 / -0.2
const FONT_BODY = {
  fontFamily: FONT_FAMILY,
  fontSize: '14px',
  lineHeight: '24px',
  letterSpacing: '-0.2px',
} as const

// `subtext` variant: 12px / 17 / -0.2
const FONT_SUBTEXT = {
  fontFamily: FONT_FAMILY,
  fontSize: '12px',
  lineHeight: '17px',
  letterSpacing: '-0.2px',
} as const

// `label` variant: 11px / 14 / -0.2 — tight, used for axis ticks.
const FONT_LABEL = {
  fontFamily: FONT_FAMILY,
  fontSize: '11px',
  lineHeight: '14px',
  letterSpacing: '-0.2px',
} as const

export const chartTooltipContentStyle: React.CSSProperties = {
  ...FONT_SUBTEXT,
  backgroundColor: 'var(--background)',
  border: '1px solid var(--border-color)',
  borderRadius: '8px',
  color: 'var(--foreground)',
  boxShadow: '0 4px 12px rgba(0, 0, 0, 0.18)',
  padding: '8px 12px',
}

export const chartTooltipLabelStyle: React.CSSProperties = {
  ...FONT_BODY,
  color: 'var(--foreground)',
  fontWeight: 500,
  marginBottom: '4px',
}

export const chartTooltipItemStyle: React.CSSProperties = {
  ...FONT_SUBTEXT,
}

export const chartAxisTickStyle = {
  ...FONT_LABEL,
  fontWeight: 400,
  fill: 'var(--foreground)',
  fillOpacity: 0.7,
} as const

export const chartGridStroke = 'var(--border-color)'

/**
 * Series colors for the policy analytics charts. We reference the design
 * system color tokens (declared in `styles.css`) so the charts stay in sync
 * with the rest of the UI — Status badges, Text themes, etc. — across both
 * light and dark mode.
 *
 * Semantic mapping mirrors the `Text` component's status themes:
 *   pass  -> green   warn -> orange   fail -> red
 */
export const CHART_SERIES_PASS = 'var(--color-green-500)'
export const CHART_SERIES_WARN = 'var(--color-orange-500)'
export const CHART_SERIES_FAIL = 'var(--color-red-500)'
