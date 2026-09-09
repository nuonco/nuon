import { DateTime } from 'luxon'
import { afterEach, describe, expect, test } from 'bun:test'
import { cleanup, render, screen } from '@testing-library/react'
import { TimelineItem } from '../../molecules/TimelineItem'
import { Timeline } from './Timeline'

type TEvent = {
  id: string
  name: string
  status: string
  createdAt: string
}

const at = (iso: string) => DateTime.fromISO(iso).toISO()

const EVENTS: TEvent[] = [
  {
    id: 'one',
    name: 'Payments API deploy',
    status: 'active',
    createdAt: at('2026-09-04T15:00:00Z'),
  },
  {
    id: 'two',
    name: 'Payments API build',
    status: 'failed',
    createdAt: at('2026-09-04T09:00:00Z'),
  },
  {
    id: 'three',
    name: 'Dashboard deploy',
    status: 'active',
    createdAt: at('2026-08-28T09:00:00Z'),
  },
]

const Example = ({
  events = EVENTS,
  loading,
  group,
}: {
  events?: TEvent[]
  loading?: boolean
  group?: boolean
}) => (
  <Timeline
    events={events}
    getEventKey={(event) => event.id}
    getEventTime={(event) => event.createdAt}
    group={group}
    loading={loading}
    emptyState="No workflows yet"
  >
    {(event) => (
      <TimelineItem
        data-event-id={event.id}
        status={event.status}
        time={event.createdAt}
        title={event.name}
      />
    )}
  </Timeline>
)

afterEach(cleanup)

describe('Timeline', () => {
  test('renders events as list items in the order given', () => {
    render(<Example />)

    expect(
      screen.getAllByRole('listitem').map((item) => item.dataset.eventId)
    ).toEqual(['one', 'two', 'three'])
  })

  test('groups events into one dated list per calendar day', () => {
    const { container } = render(<Example />)

    const lists = container.querySelectorAll('ol')
    expect(lists).toHaveLength(2)
    expect(lists[0].querySelectorAll('li')).toHaveLength(2)
    expect(lists[1].querySelectorAll('li')).toHaveLength(1)
    expect(container.querySelectorAll('h3')).toHaveLength(2)
  })

  test('renders a single flat list when grouping is off', () => {
    const { container } = render(<Example group={false} />)

    expect(container.querySelectorAll('ol')).toHaveLength(1)
    expect(container.querySelectorAll('h3')).toHaveLength(0)
    expect(screen.getAllByRole('listitem')).toHaveLength(3)
  })

  test('groups events without a usable timestamp under Undated', () => {
    render(
      <Example
        events={[
          EVENTS[0],
          { ...EVENTS[1], id: 'undated', createdAt: '' },
        ]}
      />
    )

    expect(screen.getByText('Undated')).toBeTruthy()
  })

  test('keeps repeated timestamps distinct through getEventKey', () => {
    const sameTime = EVENTS.map((event, index) => ({
      ...event,
      id: `event-${index}`,
      createdAt: EVENTS[0].createdAt,
    }))
    render(<Example events={sameTime} />)

    expect(
      screen.getAllByRole('listitem').map((item) => item.dataset.eventId)
    ).toEqual(['event-0', 'event-1', 'event-2'])
  })

  test('renders skeleton events that keep the item rhythm', () => {
    const { container } = render(<Example loading />)

    expect(screen.getByRole('status', { name: 'Loading history' })).toBeTruthy()
    expect(
      container.querySelectorAll('[data-timeline-loading-item]')
    ).toHaveLength(5)
    expect(screen.queryByText('Payments API deploy')).toBeNull()
  })

  test('renders the empty state only when a resolved page has no events', () => {
    const { rerender } = render(<Example events={[]} />)
    expect(screen.getByText('No workflows yet')).toBeTruthy()

    rerender(<Example events={[]} loading />)
    expect(screen.queryByText('No workflows yet')).toBeNull()
  })
})
