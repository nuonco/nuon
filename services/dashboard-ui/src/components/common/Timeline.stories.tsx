import { Timeline } from './Timeline'
import { TimelineEvent } from './TimelineEvent'

const events = [
  {
    created_at: '2024-07-15T12:00:00Z',
    title: 'Successful Event',
    status: 'success',
    caption: 'This event completed successfully.',
    additionalCaption: 'v1.2.3',
    created_by: 'testuser',
    badge: { children: 'Latest' },
  },
  {
    created_at: '2024-07-15T11:00:00Z',
    title: 'Failed Event',
    status: 'failed',
    caption: 'This event failed.',
    created_by: 'testuser',
    badge: { children: 'Skipped', theme: 'error' },
  },
  {
    created_at: '2024-07-15T10:00:00Z',
    title: 'Running Event',
    status: 'in-progress',
    caption: 'This event is currently running.',
    created_by: 'testuser',
  },
  {
    created_at: '2024-07-15T09:00:00Z',
    title: 'Cancelled Event',
    status: 'cancelled',
    caption: 'This event is queued.',
    created_by: 'testuser',
  },
]

export const FullTimeline = () => (
  <Timeline
    events={events}
    pagination={{
      limit: 10,
      offset: 0,
      hasNext: true,
    }}
    renderEvent={(event) => (
      <TimelineEvent
        key={event.status}
        createdAt={event.created_at}
        caption={event.caption}
        createdBy={event.created_by}
        status={event.status}
        title={event.title}
        additionalCaption={event.additionalCaption}
      />
    )}
  />
)
