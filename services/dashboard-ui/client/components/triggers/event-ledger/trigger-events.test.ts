import { describe, expect, test } from 'bun:test'
import { triggerEventsPath } from '@/lib/ctl-api/triggers'

describe('triggerEventsPath', () => {
  test('constructs the default trigger event request', () => {
    expect(triggerEventsPath('trigger/one', {})).toBe(
      'triggers/trigger%2Fone/events?limit=100&order=desc'
    )
  })

  test('constructs filtered ascending requests', () => {
    expect(
      triggerEventsPath('trigger-one', {
        cursor: 'next page',
        eventType: 'push/created',
        limit: 25,
        order: 'asc',
        outcome: 'failed',
        query: 'external id',
      })
    ).toBe(
      'triggers/trigger-one/events?limit=25&order=asc&cursor=next+page&event_type=push%2Fcreated&outcome=failed&query=external+id'
    )
  })
})
