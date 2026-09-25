import { expect, test } from 'bun:test'
import { isXYChart, parseXYChart } from './parse-xychart'

test('detects xychart headers', () => {
  expect(isXYChart('xychart-beta\n  line [1, 2]')).toBe(true)
  expect(isXYChart('xychart\n  line [1, 2]')).toBe(true)
  expect(isXYChart('graph TD\n  A --> B')).toBe(false)
})

test('parses the docs sample', () => {
  const chart = parseXYChart(`xychart-beta
    title "Product Sales (2026)"
    x-axis [Jan, Feb, Mar, Apr, May, Jun]
    y-axis "Revenue (in USD)" 0 --> 5000
    bar [1500, 2500, 3200, 2800, 4200, 4800]
    line [1500, 2500, 3200, 2800, 4200, 4800]`)

  expect(chart).not.toBeNull()
  expect(chart?.title).toBe('Product Sales (2026)')
  expect(chart?.xLabels).toEqual(['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun'])
  expect(chart?.yLabel).toBe('Revenue (in USD)')
  expect(chart?.yMin).toBe(0)
  expect(chart?.yMax).toBe(5000)
  expect(chart?.series).toHaveLength(2)
  expect(chart?.series[0]).toEqual({ kind: 'bar', values: [1500, 2500, 3200, 2800, 4200, 4800] })
  expect(chart?.series[1]?.kind).toBe('line')
})

test('parses a minimal runbook chart', () => {
  const chart = parseXYChart(`xychart-beta
    title "CPU %"
    line [12.62, 14.19, 12.75]`)

  expect(chart?.series).toEqual([{ kind: 'line', values: [12.62, 14.19, 12.75] }])
  expect(chart?.xLabels).toBeUndefined()
})

test('rejects non-numeric series and unknown directives', () => {
  expect(parseXYChart('xychart-beta\n  line [1, oops, 3]')).toBeNull()
  expect(parseXYChart('xychart-beta\n  pie [1, 2]')).toBeNull()
  expect(parseXYChart('xychart-beta\n  title "empty"')).toBeNull()
})
