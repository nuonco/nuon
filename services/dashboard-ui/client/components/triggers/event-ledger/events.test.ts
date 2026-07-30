import { describe, expect, test } from 'bun:test'
import {
  EVENT_PATH_TRUNCATION_MARKER,
  eventOutcome,
  flattenEventPaths,
  isTriggerEventTerminal,
  markTriggerEventDispatchPending,
  shouldPollTriggerEventDispatches,
} from './events'

describe('flattenEventPaths', () => {
  test('renders deterministic RFC 9535 paths and preserves empty containers', () => {
    expect(
      flattenEventPaths({
        empty: {},
        items: [true],
        'bad key': { 'line\nbreak': null },
      })
    ).toEqual([
      { path: '$["bad key"]["line\\nbreak"]', value: null },
      { path: '$.empty', value: {} },
      { path: '$.items[0]', value: true },
    ])
  })

  test('limits paths and includes a visible truncation marker', () => {
    expect(flattenEventPaths({ a: 1, b: 2, c: 3 }, 20, 2)).toEqual([
      { path: '$.a', value: 1 },
      { path: '$.b', value: 2 },
      { path: '…', value: EVENT_PATH_TRUNCATION_MARKER },
    ])
  })

  test('limits depth without recursive traversal', () => {
    let payload: unknown = 'leaf'
    for (let index = 0; index < 10_000; index += 1) payload = { child: payload }

    expect(flattenEventPaths(payload, 5)).toEqual([
      { path: '…', value: EVENT_PATH_TRUNCATION_MARKER },
    ])
  })

  test('does not enqueue an entire wide array', () => {
    expect(
      flattenEventPaths(
        Array.from({ length: 10_000 }, (_, index) => index),
        20,
        3
      )
    ).toEqual([
      { path: '$[0]', value: 0 },
      { path: '$[1]', value: 1 },
      { path: '$[2]', value: 2 },
      { path: '…', value: EVENT_PATH_TRUNCATION_MARKER },
    ])
  })
})

describe('isTriggerEventTerminal', () => {
  test('requires terminal routing and dispatch statuses', () => {
    expect(isTriggerEventTerminal({ routing_status: 'routing' })).toBe(false)
    expect(
      isTriggerEventTerminal({
        routing_status: 'matched',
        dispatches: [{ status: 'dispatching' }],
      })
    ).toBe(false)
    expect(
      isTriggerEventTerminal({
        routing_status: 'matched',
        dispatches: [{ status: 'triggered' }, { status: 'dead_lettered' }],
      })
    ).toBe(true)
    expect(isTriggerEventTerminal({ routing_status: 'ignored' })).toBe(true)
    expect(
      isTriggerEventTerminal({
        routing_status: 'matched',
        dispatches_truncated: true,
        dispatches: [{ status: 'triggered' }],
      })
    ).toBe(false)
  })
})

describe('trigger event dispatch polling', () => {
  test('keeps polling an empty result while routing and after a dispatch appears', () => {
    expect(
      shouldPollTriggerEventDispatches({
        event: { routing_status: 'routing', dispatch_count: 0 },
        pages: [{ items: [] }],
      })
    ).toBe(true)
    expect(
      shouldPollTriggerEventDispatches({
        event: { routing_status: 'matched', dispatch_count: 1 },
        pages: [{ items: [{ id: 'dispatch-1', status: 'dispatching' }] }],
      })
    ).toBe(true)
  })

  test('marks a dead-lettered retry pending so polling resumes', () => {
    const pages = markTriggerEventDispatchPending(
      [{ items: [{ id: 'dispatch-1', status: 'dead_lettered' }] }],
      'dispatch-1'
    )
    expect(pages[0]?.items?.[0]?.status).toBe('pending')
    expect(
      shouldPollTriggerEventDispatches({
        event: { routing_status: 'matched', dispatch_count: 1 },
        pages,
      })
    ).toBe(true)
  })
})

describe('eventOutcome', () => {
  test('maps routing state to an event outcome', () => {
    expect(eventOutcome({ routing_status: 'matched' })).toBe('ok')
    expect(eventOutcome({ routing_status: 'ignored' })).toBe('ignored')
    expect(eventOutcome({ routing_status: 'rejected' })).toBe('rejected')
    expect(eventOutcome({ routing_status: 'routing_failed' })).toBe('failed')
    expect(eventOutcome({ routing_status: 'routing' })).toBe('processing')
  })
})
