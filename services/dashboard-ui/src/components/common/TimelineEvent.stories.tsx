import { TimelineEvent } from './TimelineEvent'

export const Default = () => (
  <TimelineEvent
    title="Default Event"
    status="default"
    createdAt="2024-07-15T12:00:00Z"
    caption="This is a default timeline event"
  />
)

export const AllStatuses = () => (
  <div className="flex flex-col">
    <TimelineEvent
      title="Success Event"
      status="success"
      createdAt="2024-07-15T12:00:00Z"
      caption="This event completed successfully"
      createdBy="testuser"
    />
    <TimelineEvent
      title="Failed Event"
      status="failed"
      createdAt="2024-07-15T11:00:00Z"
      caption="This event failed"
      createdBy="testuser"
    />
    <TimelineEvent
      title="In Progress Event"
      status="in-progress"
      createdAt="2024-07-15T10:00:00Z"
      caption="This event is currently running"
      createdBy="testuser"
    />
    <TimelineEvent
      title="Cancelled Event"
      status="cancelled"
      createdAt="2024-07-15T09:00:00Z"
      caption="This event was cancelled"
      createdBy="testuser"
    />
    <TimelineEvent
      title="Warning Event"
      status="warn"
      createdAt="2024-07-15T08:00:00Z"
      caption="This event has a warning"
      createdBy="testuser"
    />
  </div>
)

export const WithBadge = () => (
  <div className="flex flex-col">
    <TimelineEvent
      title="Event with Badge"
      status="success"
      createdAt="2024-07-15T12:00:00Z"
      caption="This event has a badge"
      createdBy="testuser"
      badge={{ children: 'Latest' }}
    />
    <TimelineEvent
      title="Event with Error Badge"
      status="failed"
      createdAt="2024-07-15T11:00:00Z"
      caption="This event has an error badge"
      createdBy="testuser"
      badge={{ children: 'Skipped', theme: 'error' }}
    />
    <TimelineEvent
      title="Event with Warning Badge"
      status="warn"
      createdAt="2024-07-15T10:00:00Z"
      caption="This event has a warning badge"
      createdBy="testuser"
      badge={{ children: 'Retry', theme: 'warn' }}
    />
  </div>
)

export const WithAdditionalCaption = () => (
  <div className="flex flex-col">
    <TimelineEvent
      title="Event with Version"
      status="success"
      createdAt="2024-07-15T12:00:00Z"
      caption="Deployment successful"
      additionalCaption="v1.2.3"
      createdBy="testuser"
    />
    <TimelineEvent
      title="Event with Build Info"
      status="success"
      createdAt="2024-07-15T11:00:00Z"
      caption="Build completed"
      additionalCaption="build-456"
      createdBy="testuser"
    />
  </div>
)

export const WithComplexTitle = () => (
  <TimelineEvent
    title={
      <span>
        Complex title with <strong>bold text</strong> and{' '}
        <code className="bg-gray-100 px-1 rounded">code</code>
      </span>
    }
    status="success"
    createdAt="2024-07-15T12:00:00Z"
    caption="This event has a complex title with React nodes"
    createdBy="testuser"
  />
)

export const WithoutOptionalFields = () => (
  <TimelineEvent
    title="Minimal Event"
    status="default"
    createdAt="2024-07-15T12:00:00Z"
  />
)

export const FullExample = () => (
  <div className="flex flex-col">
    <TimelineEvent
      title="Complete deployment"
      status="success"
      createdAt="2024-07-15T12:00:00Z"
      caption="Successfully deployed to production"
      additionalCaption="v2.1.0"
      createdBy="john.doe"
      badge={{ children: 'Latest', theme: 'success' }}
    />
    <TimelineEvent
      title="Build failed"
      status="failed"
      createdAt="2024-07-15T11:30:00Z"
      caption="Build failed due to test failures"
      additionalCaption="build-789"
      createdBy="jane.smith"
      badge={{ children: 'Failed', theme: 'error' }}
    />
    <TimelineEvent
      title="Starting build"
      status="in-progress"
      createdAt="2024-07-15T11:00:00Z"
      caption="Build started for staging environment"
      additionalCaption="build-788"
      createdBy="bot"
    />
  </div>
)
