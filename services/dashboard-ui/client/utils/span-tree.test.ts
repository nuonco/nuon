import { describe, expect, test } from 'bun:test'
import type { TSpan } from '@/types'
import {
  buildSpanForest,
  collectDescendantIds,
  flattenForest,
  formatDurationNs,
  traceEnd,
  traceStart,
} from './span-tree'

const span = (overrides: Partial<TSpan> & { span_id: string }): TSpan => ({
  name: overrides.span_id,
  start_time: '2024-01-01T00:00:00.000Z',
  end_time: '2024-01-01T00:00:01.000Z',
  duration_ns: 1_000_000_000,
  ...overrides,
})

describe('span-tree', () => {
  describe('buildSpanForest', () => {
    test('returns empty array for empty input', () => {
      expect(buildSpanForest([])).toEqual([])
    })

    test('returns empty array for null-ish input', () => {
      expect(buildSpanForest(null as any)).toEqual([])
      expect(buildSpanForest(undefined as any)).toEqual([])
    })

    test('single span becomes a root', () => {
      const spans = [span({ span_id: 'a' })]
      const forest = buildSpanForest(spans)
      expect(forest).toHaveLength(1)
      expect(forest[0].span.span_id).toBe('a')
      expect(forest[0].depth).toBe(0)
      expect(forest[0].children).toHaveLength(0)
    })

    test('child span is nested under parent', () => {
      const spans = [
        span({ span_id: 'parent', start_time: '2024-01-01T00:00:00.000Z' }),
        span({ span_id: 'child', parent_span_id: 'parent', start_time: '2024-01-01T00:00:00.100Z' }),
      ]
      const forest = buildSpanForest(spans)
      expect(forest).toHaveLength(1)
      expect(forest[0].span.span_id).toBe('parent')
      expect(forest[0].children).toHaveLength(1)
      expect(forest[0].children[0].span.span_id).toBe('child')
      expect(forest[0].children[0].depth).toBe(1)
    })

    test('siblings are sorted by start_time', () => {
      const spans = [
        span({ span_id: 'root' }),
        span({ span_id: 'b', parent_span_id: 'root', start_time: '2024-01-01T00:00:02.000Z' }),
        span({ span_id: 'a', parent_span_id: 'root', start_time: '2024-01-01T00:00:01.000Z' }),
      ]
      const forest = buildSpanForest(spans)
      expect(forest[0].children[0].span.span_id).toBe('a')
      expect(forest[0].children[1].span.span_id).toBe('b')
    })

    test('multiple roots are sorted by start_time', () => {
      const spans = [
        span({ span_id: 'second', start_time: '2024-01-01T00:00:02.000Z' }),
        span({ span_id: 'first', start_time: '2024-01-01T00:00:01.000Z' }),
      ]
      const forest = buildSpanForest(spans)
      expect(forest[0].span.span_id).toBe('first')
      expect(forest[1].span.span_id).toBe('second')
    })

    test('orphaned parent_span_id becomes a root', () => {
      const spans = [
        span({ span_id: 'child', parent_span_id: 'missing' }),
      ]
      const forest = buildSpanForest(spans)
      expect(forest).toHaveLength(1)
      expect(forest[0].span.span_id).toBe('child')
      expect(forest[0].depth).toBe(0)
    })

    test('hasErrorDescendant bubbles up from leaf error', () => {
      const spans = [
        span({ span_id: 'root' }),
        span({ span_id: 'mid', parent_span_id: 'root' }),
        span({ span_id: 'leaf', parent_span_id: 'mid', status_code: 'Error' }),
      ]
      const forest = buildSpanForest(spans)
      expect(forest[0].hasErrorDescendant).toBe(true)
      expect(forest[0].children[0].hasErrorDescendant).toBe(true)
      expect(forest[0].children[0].children[0].hasErrorDescendant).toBe(true)
    })

    test('hasErrorDescendant is false when no errors', () => {
      const spans = [
        span({ span_id: 'root' }),
        span({ span_id: 'child', parent_span_id: 'root', status_code: 'Ok' }),
      ]
      const forest = buildSpanForest(spans)
      expect(forest[0].hasErrorDescendant).toBe(false)
    })

    test('depth is set correctly for deep trees', () => {
      const spans = [
        span({ span_id: 'a' }),
        span({ span_id: 'b', parent_span_id: 'a' }),
        span({ span_id: 'c', parent_span_id: 'b' }),
        span({ span_id: 'd', parent_span_id: 'c' }),
      ]
      const forest = buildSpanForest(spans)
      const flat = flattenForest(forest)
      expect(flat.map((n) => n.depth)).toEqual([0, 1, 2, 3])
    })
  })

  describe('collectDescendantIds', () => {
    const spans = [
      span({ span_id: 'root' }),
      span({ span_id: 'a', parent_span_id: 'root' }),
      span({ span_id: 'b', parent_span_id: 'root' }),
      span({ span_id: 'a1', parent_span_id: 'a' }),
      span({ span_id: 'a2', parent_span_id: 'a' }),
    ]

    test('collects all descendants of root', () => {
      const ids = collectDescendantIds(spans, 'root')
      expect(ids).toEqual(new Set(['root', 'a', 'b', 'a1', 'a2']))
    })

    test('collects descendants of intermediate node', () => {
      const ids = collectDescendantIds(spans, 'a')
      expect(ids).toEqual(new Set(['a', 'a1', 'a2']))
    })

    test('leaf returns just itself', () => {
      const ids = collectDescendantIds(spans, 'a1')
      expect(ids).toEqual(new Set(['a1']))
    })

    test('missing span returns just the id', () => {
      const ids = collectDescendantIds(spans, 'nonexistent')
      expect(ids).toEqual(new Set(['nonexistent']))
    })

    test('empty spans returns just the id', () => {
      const ids = collectDescendantIds([], 'x')
      expect(ids).toEqual(new Set(['x']))
    })

    test('empty spanId returns just the empty string', () => {
      const ids = collectDescendantIds(spans, '')
      expect(ids).toEqual(new Set(['']))
    })
  })

  describe('flattenForest', () => {
    test('returns empty for empty forest', () => {
      expect(flattenForest([])).toEqual([])
    })

    test('flattens depth-first', () => {
      const spans = [
        span({ span_id: 'root', start_time: '2024-01-01T00:00:00.000Z' }),
        span({ span_id: 'a', parent_span_id: 'root', start_time: '2024-01-01T00:00:01.000Z' }),
        span({ span_id: 'a1', parent_span_id: 'a', start_time: '2024-01-01T00:00:02.000Z' }),
        span({ span_id: 'b', parent_span_id: 'root', start_time: '2024-01-01T00:00:03.000Z' }),
      ]
      const forest = buildSpanForest(spans)
      const flat = flattenForest(forest)
      expect(flat.map((n) => n.span.span_id)).toEqual(['root', 'a', 'a1', 'b'])
    })
  })

  describe('traceStart', () => {
    test('returns 0 for empty input', () => {
      expect(traceStart([])).toBe(0)
    })

    test('returns earliest start_time in epoch ms', () => {
      const spans = [
        span({ span_id: 'a', start_time: '2024-01-01T00:00:05.000Z' }),
        span({ span_id: 'b', start_time: '2024-01-01T00:00:01.000Z' }),
        span({ span_id: 'c', start_time: '2024-01-01T00:00:10.000Z' }),
      ]
      expect(traceStart(spans)).toBe(new Date('2024-01-01T00:00:01.000Z').getTime())
    })
  })

  describe('traceEnd', () => {
    test('returns 0 for empty input', () => {
      expect(traceEnd([])).toBe(0)
    })

    test('returns latest end_time in epoch ms', () => {
      const spans = [
        span({ span_id: 'a', end_time: '2024-01-01T00:00:05.000Z' }),
        span({ span_id: 'b', end_time: '2024-01-01T00:00:10.000Z' }),
        span({ span_id: 'c', end_time: '2024-01-01T00:00:01.000Z' }),
      ]
      expect(traceEnd(spans)).toBe(new Date('2024-01-01T00:00:10.000Z').getTime())
    })
  })

  describe('formatDurationNs', () => {
    test('returns dash for non-finite values', () => {
      expect(formatDurationNs(NaN)).toBe('—')
      expect(formatDurationNs(Infinity)).toBe('—')
      expect(formatDurationNs(-Infinity)).toBe('—')
    })

    test('returns dash for zero and negative', () => {
      expect(formatDurationNs(0)).toBe('—')
      expect(formatDurationNs(-100)).toBe('—')
    })

    test('formats sub-millisecond as nanoseconds', () => {
      expect(formatDurationNs(500)).toBe('500 ns')
      expect(formatDurationNs(999_999)).toBe('999999 ns')
    })

    test('formats milliseconds', () => {
      expect(formatDurationNs(1_000_000)).toBe('1.0 ms')
      expect(formatDurationNs(500_000_000)).toBe('500.0 ms')
    })

    test('formats seconds', () => {
      expect(formatDurationNs(1_000_000_000)).toBe('1.00 s')
      expect(formatDurationNs(2_500_000_000)).toBe('2.50 s')
    })

    test('formats minutes and seconds', () => {
      expect(formatDurationNs(60_000_000_000)).toBe('1m 0s')
      expect(formatDurationNs(90_000_000_000)).toBe('1m 30s')
      expect(formatDurationNs(125_000_000_000)).toBe('2m 5s')
    })
  })
})
