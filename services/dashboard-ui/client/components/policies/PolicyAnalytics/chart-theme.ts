const FONT_FAMILY =
  'var(--font-inter), -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif'

const FONT_BODY = {
  fontFamily: FONT_FAMILY,
  fontSize: '14px',
  lineHeight: '24px',
  letterSpacing: '-0.2px',
} as const

const FONT_SUBTEXT = {
  fontFamily: FONT_FAMILY,
  fontSize: '12px',
  lineHeight: '17px',
  letterSpacing: '-0.2px',
} as const

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

export const CHART_SERIES_PASS = 'var(--color-green-500)'
export const CHART_SERIES_WARN = 'var(--color-orange-500)'
export const CHART_SERIES_FAIL = 'var(--color-red-500)'
