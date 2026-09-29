/**
 * Parser for mermaid `xychart-beta` blocks — the subset used by runbook
 * readmes: an optional title, optional axes, and any number of line/bar
 * series. Returns null when the block doesn't parse, so the markdown
 * renderer can fall back to showing the source in a code block.
 *
 *   xychart-beta
 *       title "CPU %"
 *       x-axis [Jan, Feb, Mar]          (or: x-axis "label" 1 --> 60)
 *       y-axis "Revenue" 0 --> 5000     (or: y-axis "label")
 *       bar [1500, 2500, 3200]
 *       line [1500, 2500, 3200]
 */

export type ParsedXYSeries = {
  kind: 'line' | 'bar'
  values: number[]
}

export type ParsedXYChart = {
  title?: string
  xLabels?: string[]
  yLabel?: string
  yMin?: number
  yMax?: number
  series: ParsedXYSeries[]
}

export const isXYChart = (code: string) => /^xychart(?:-beta)?\s*(?:horizontal\s*)?$/i.test(code.trim().split('\n')[0] ?? '')

const parseNumberList = (raw: string): number[] | null => {
  const values: number[] = []
  for (const part of raw.split(',')) {
    const n = Number(part.trim())
    if (!Number.isFinite(n)) return null
    values.push(n)
  }
  return values
}

export function parseXYChart(code: string): ParsedXYChart | null {
  const lines = code
    .trim()
    .split('\n')
    .map((l) => l.trim())
    .filter((l) => l.length > 0)

  if (lines.length < 2 || !isXYChart(lines[0])) return null

  const chart: ParsedXYChart = { series: [] }

  for (const line of lines.slice(1)) {
    const titleMatch = line.match(/^title\s+"([^"]*)"$/)
    if (titleMatch) {
      chart.title = titleMatch[1]
      continue
    }

    const xCategories = line.match(/^x-axis\s+(?:"[^"]*"\s+)?\[(.*)\]$/)
    if (xCategories) {
      chart.xLabels = xCategories[1].split(',').map((s) => s.trim())
      continue
    }

    // x-axis numeric range — nothing to keep, indices work fine
    if (/^x-axis\b/.test(line)) continue

    const yAxis = line.match(/^y-axis\s+(?:"([^"]*)")?\s*(?:(-?[\d.]+)\s*-->\s*(-?[\d.]+))?$/)
    if (yAxis) {
      if (yAxis[1]) chart.yLabel = yAxis[1]
      if (yAxis[2] !== undefined && yAxis[3] !== undefined) {
        chart.yMin = Number(yAxis[2])
        chart.yMax = Number(yAxis[3])
      }
      continue
    }

    const series = line.match(/^(line|bar)\s+(?:"[^"]*"\s+)?\[(.*)\]$/)
    if (series) {
      const values = parseNumberList(series[2])
      if (!values || values.length === 0) return null
      chart.series.push({ kind: series[1] as 'line' | 'bar', values })
      continue
    }

    // unknown directive — bail so the source stays visible instead of a wrong chart
    return null
  }

  if (chart.series.length === 0) return null
  return chart
}
