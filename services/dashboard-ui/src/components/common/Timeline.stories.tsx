/* eslint-disable react/no-unescaped-entities */
import { Timeline } from './Timeline'
import { TimelineEvent } from './TimelineEvent'
import { Text } from './Text'
import { type IBadge } from './Badge'
import { Button } from './Button'

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

export const BasicUsage = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Basic Timeline Usage</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Timeline displays a chronological sequence of events with status
        indicators and detailed information. It supports pagination for large
        event lists and uses TimelineEvent components for consistent event
        rendering with dates, statuses, and contextual information.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Event Timeline</h4>
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
      <Text variant="subtext" theme="neutral">
        Timeline shows events in reverse chronological order with status
        indicators
      </Text>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Key Features:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>
          Chronological event display with visual timeline connecting events
        </li>
        <li>Status-based styling with color-coded indicators</li>
        <li>Pagination support for large event sequences</li>
        <li>Flexible event rendering with custom components</li>
        <li>Responsive design that works on mobile and desktop</li>
      </ul>
    </div>
  </div>
)

export const DifferentEventTypes = () => {
  const deploymentEvents = [
    {
      created_at: '2024-07-15T12:00:00Z',
      title: 'Deployment Successful',
      status: 'success',
      caption: 'Application deployed to production environment',
      additionalCaption: 'v2.1.0',
      created_by: 'deploy-bot',
      badge: { children: 'Production', theme: 'success' },
    },
    {
      created_at: '2024-07-15T11:45:00Z',
      title: 'Tests Passed',
      status: 'success',
      caption: 'All integration tests completed successfully',
      additionalCaption: '156/156 tests passed',
      created_by: 'ci-system',
    },
    {
      created_at: '2024-07-15T11:30:00Z',
      title: 'Build Failed',
      status: 'failed',
      caption: 'Compilation failed due to TypeScript errors',
      created_by: 'ci-system',
      badge: { children: 'Retry', theme: 'error' },
    },
    {
      created_at: '2024-07-15T11:15:00Z',
      title: 'Build Started',
      status: 'in-progress',
      caption: 'Building application from source code',
      created_by: 'developer',
    },
    {
      created_at: '2024-07-15T11:00:00Z',
      title: 'Code Committed',
      status: 'success',
      caption: 'New features and bug fixes committed to main branch',
      additionalCaption: 'commit abc123f',
      created_by: 'developer',
    },
  ]

  return (
    <div className="space-y-6">
      <div className="space-y-3">
        <h3 className="text-lg font-semibold">Different Event Types</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          Timeline can display various types of events including deployments,
          builds, tests, and system activities. Each event type can have
          different statuses, badges, and contextual information.
        </p>
      </div>

      <div className="space-y-4">
        <h4 className="text-sm font-medium">Deployment Pipeline Timeline</h4>
        <Timeline
          events={deploymentEvents}
          pagination={{
            limit: 10,
            offset: 0,
            hasNext: false,
          }}
          renderEvent={(event) => (
            <TimelineEvent
              key={`${event.created_at}-${event.title}`}
              createdAt={event.created_at}
              caption={event.caption}
              createdBy={event.created_by}
              status={event.status}
              title={event.title}
              additionalCaption={event.additionalCaption}
              badge={event.badge as IBadge}
            />
          )}
        />
      </div>
    </div>
  )
}

export const PaginatedTimeline = () => {
  const manyEvents = Array.from({ length: 15 }, (_, i) => ({
    created_at: new Date(Date.now() - i * 3600000).toISOString(),
    title: `Event ${15 - i}`,
    status: ['success', 'failed', 'in-progress', 'cancelled'][i % 4],
    caption: `This is event number ${15 - i} with some descriptive information.`,
    created_by: ['alice', 'bob', 'charlie', 'system'][i % 4],
    additionalCaption: i % 3 === 0 ? `v1.${15 - i}.0` : undefined,
  }))

  return (
    <div className="space-y-6">
      <div className="space-y-3">
        <h3 className="text-lg font-semibold">Paginated Timeline</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          For large event lists, Timeline supports pagination to improve
          performance and user experience. Users can load more events as needed
          using the "Load More" functionality.
        </p>
      </div>

      <div className="space-y-4">
        <h4 className="text-sm font-medium">
          Large Event History (15 events, showing first 8)
        </h4>
        <Timeline
          events={manyEvents.slice(0, 8)}
          pagination={{
            limit: 8,
            offset: 0,
            hasNext: true,
          }}
          renderEvent={(event) => (
            <TimelineEvent
              key={`${event.created_at}-${event.title}`}
              createdAt={event.created_at}
              caption={event.caption}
              createdBy={event.created_by}
              status={event.status}
              title={event.title}
              additionalCaption={event.additionalCaption}
            />
          )}
        />
        <Text variant="subtext" theme="neutral">
          "Load More" button appears when there are additional events to display
        </Text>
      </div>
    </div>
  )
}

export const CustomEventRendering = () => {
  const customEvents = [
    {
      created_at: '2024-07-15T14:00:00Z',
      title: 'User Registration',
      status: 'success',
      caption: 'New user account created',
      created_by: 'auth-service',
      user: { name: 'John Doe', email: 'john@example.com' },
    },
    {
      created_at: '2024-07-15T13:30:00Z',
      title: 'Payment Processed',
      status: 'success',
      caption: 'Monthly subscription payment completed',
      created_by: 'payment-system',
      payment: { amount: '$29.99', method: 'Credit Card' },
    },
    {
      created_at: '2024-07-15T13:00:00Z',
      title: 'API Rate Limit Exceeded',
      status: 'failed',
      caption: 'Client exceeded API rate limits',
      created_by: 'api-gateway',
      api: { endpoint: '/api/v1/users', limit: '1000/hour' },
    },
  ]

  const renderCustomEvent = (event) => (
    <div key={`${event.created_at}-${event.title}`} className="timeline-event">
      <TimelineEvent
        createdAt={event.created_at}
        caption={event.caption}
        createdBy={event.created_by}
        status={event.status}
        title={event.title}
      />
      {/* Custom additional content */}
      {event.user && (
        <div className="ml-8 mt-2 p-3 bg-gray-50 dark:bg-gray-800 rounded border-l-2 border-blue-200 dark:border-blue-800">
          <Text variant="subtext" weight="strong">
            User Details:
          </Text>
          <Text variant="subtext">
            {event.user.name} ({event.user.email})
          </Text>
        </div>
      )}
      {event.payment && (
        <div className="ml-8 mt-2 p-3 bg-gray-50 dark:bg-gray-800 rounded border-l-2 border-green-200 dark:border-green-800">
          <Text variant="subtext" weight="strong">
            Payment Details:
          </Text>
          <Text variant="subtext">
            {event.payment.amount} via {event.payment.method}
          </Text>
        </div>
      )}
      {event.api && (
        <div className="ml-8 mt-2 p-3 bg-gray-50 dark:bg-gray-800 rounded border-l-2 border-red-200 dark:border-red-800">
          <Text variant="subtext" weight="strong">
            API Details:
          </Text>
          <Text variant="subtext">
            {event.api.endpoint} (Limit: {event.api.limit})
          </Text>
        </div>
      )}
    </div>
  )

  return (
    <div className="space-y-6">
      <div className="space-y-3">
        <h3 className="text-lg font-semibold">Custom Event Rendering</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          Timeline supports custom event rendering for specialized use cases.
          You can add additional content, custom styling, and context-specific
          information to timeline events.
        </p>
      </div>

      <div className="space-y-4">
        <h4 className="text-sm font-medium">Enhanced Event Details</h4>
        <Timeline
          events={customEvents}
          pagination={{
            limit: 10,
            offset: 0,
            hasNext: false,
          }}
          renderEvent={renderCustomEvent}
        />
      </div>
    </div>
  )
}

export const UsageExamples = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Common Usage Patterns</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Timelines are commonly used for activity feeds, deployment histories,
        audit logs, and process tracking. They provide chronological context and
        visual progression of events over time.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">System Activity Timeline</h4>
      <div className="p-4 border rounded-lg">
        <div className="flex justify-between items-center mb-4">
          <Text variant="h3" weight="stronger">
            Recent Activity
          </Text>
          <Button variant="ghost" size="sm">
            View All
          </Button>
        </div>
        <Timeline
          events={events.slice(0, 3)}
          pagination={{
            limit: 3,
            offset: 0,
            hasNext: true,
          }}
          renderEvent={(event) => (
            <TimelineEvent
              key={`${event.created_at}-${event.title}`}
              createdAt={event.created_at}
              caption={event.caption}
              createdBy={event.created_by}
              status={event.status}
              title={event.title}
              additionalCaption={event.additionalCaption}
              badge={event.badge as IBadge}
            />
          )}
        />
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Best Practices:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>
          Use consistent event titles and descriptions for similar activities
        </li>
        <li>
          Include relevant contextual information like version numbers or user
          names
        </li>
        <li>
          Choose appropriate status indicators that match your application's
          conventions
        </li>
        <li>Implement pagination for timelines with many events</li>
        <li>Consider using badges for additional categorization or priority</li>
        <li>Provide "created by" information for accountability and context</li>
        <li>
          Use relative timestamps for recent events and absolute dates for
          historical ones
        </li>
      </ul>
    </div>
  </div>
)
