/* eslint-disable react/no-unescaped-entities */
import { TimelineEvent } from './TimelineEvent'
import { Text } from './Text'
import { Badge } from './Badge'
import { Button } from './Button'

export const BasicUsage = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Basic TimelineEvent Usage</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        TimelineEvent represents a single event in a chronological sequence.
        It includes a status indicator, timestamp, title, caption, and
        optional metadata like creator information and badges. Events
        connect visually to form complete timelines.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Simple Event</h4>
      <div className="border rounded-lg p-4">
        <TimelineEvent
          title="Application Deployed"
          status="success"
          createdAt="2024-07-15T12:00:00Z"
          caption="Successfully deployed to production environment"
          createdBy="deploy-bot"
        />
      </div>
      <Text variant="subtext" theme="neutral">
        Basic event with title, status, timestamp, and description
      </Text>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Key Features:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Status-based visual indicators with color coding</li>
        <li>Flexible title support including React components</li>
        <li>Automatic timestamp formatting and display</li>
        <li>Optional badges for additional categorization</li>
        <li>Creator attribution with "created by" information</li>
        <li>Additional caption support for version numbers or IDs</li>
      </ul>
    </div>
  </div>
)

export const StatusVariants = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Status Variants</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        TimelineEvent supports multiple status types with distinct visual
        indicators. Each status uses appropriate colors and styling to
        communicate the event outcome or current state.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">All Status Types</h4>
      <div className="space-y-0 border rounded-lg p-4">
        <TimelineEvent
          title="Successful Deployment"
          status="success"
          createdAt="2024-07-15T12:00:00Z"
          caption="Application deployed successfully to production"
          createdBy="deploy-system"
        />
        <TimelineEvent
          title="Build Failed"
          status="failed"
          createdAt="2024-07-15T11:00:00Z"
          caption="Compilation failed due to TypeScript errors"
          createdBy="ci-system"
        />
        <TimelineEvent
          title="Currently Running"
          status="in-progress"
          createdAt="2024-07-15T10:00:00Z"
          caption="Integration tests are currently executing"
          createdBy="test-runner"
        />
        <TimelineEvent
          title="Deployment Cancelled"
          status="cancelled"
          createdAt="2024-07-15T09:00:00Z"
          caption="User cancelled deployment before completion"
          createdBy="developer"
        />
        <TimelineEvent
          title="Warning Detected"
          status="warn"
          createdAt="2024-07-15T08:00:00Z"
          caption="High memory usage detected during deployment"
          createdBy="monitoring-system"
        />
      </div>
      <Text variant="subtext" theme="neutral">
        Each status type has distinctive colors and visual indicators
      </Text>
    </div>
  </div>
)

export const WithBadges = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Events with Badges</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Timeline events can include badges for additional categorization,
        priority indication, or status clarification. Badges use the same
        theming system as other components for consistency.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Badge Examples</h4>
      <div className="space-y-0 border rounded-lg p-4">
        <TimelineEvent
          title="Production Release"
          status="success"
          createdAt="2024-07-15T12:00:00Z"
          caption="New version deployed to production environment"
          createdBy="release-manager"
          badge={{ children: 'Latest', theme: 'brand' }}
        />
        <TimelineEvent
          title="Failed Integration Test"
          status="failed"
          createdAt="2024-07-15T11:00:00Z"
          caption="Test suite failed and was automatically skipped"
          createdBy="ci-pipeline"
          badge={{ children: 'Skipped', theme: 'error' }}
        />
        <TimelineEvent
          title="Retry Scheduled"
          status="warn"
          createdAt="2024-07-15T10:00:00Z"
          caption="Build failed but automatic retry is scheduled"
          createdBy="build-system"
          badge={{ children: 'Auto-Retry', theme: 'warn' }}
        />
        <TimelineEvent
          title="Manual Intervention"
          status="in-progress"
          createdAt="2024-07-15T09:00:00Z"
          caption="Deployment requires manual approval to proceed"
          createdBy="security-check"
          badge={{ children: 'Approval Required', theme: 'info' }}
        />
      </div>
    </div>
  </div>
)

export const WithAdditionalInfo = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Additional Information</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Timeline events can include additional captions for version numbers,
        build IDs, commit hashes, or other relevant metadata. This provides
        extra context without cluttering the main event description.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Metadata Examples</h4>
      <div className="space-y-0 border rounded-lg p-4">
        <TimelineEvent
          title="Version Release"
          status="success"
          createdAt="2024-07-15T12:00:00Z"
          caption="Successfully released new application version"
          additionalCaption="v2.1.0"
          createdBy="release-bot"
        />
        <TimelineEvent
          title="Build Completed"
          status="success"
          createdAt="2024-07-15T11:00:00Z"
          caption="Application built successfully from source"
          additionalCaption="build-456789"
          createdBy="ci-system"
        />
        <TimelineEvent
          title="Code Committed"
          status="success"
          createdAt="2024-07-15T10:00:00Z"
          caption="New features and bug fixes committed to main branch"
          additionalCaption="commit abc123f"
          createdBy="developer"
        />
        <TimelineEvent
          title="Environment Updated"
          status="success"
          createdAt="2024-07-15T09:00:00Z"
          caption="Configuration updated for staging environment"
          additionalCaption="config-v3.2"
          createdBy="devops-team"
        />
      </div>
    </div>
  </div>
)

export const FlexibleTitles = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Flexible Title Content</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Timeline event titles can be simple strings or complex React
        components with formatting, links, and interactive elements.
        This flexibility allows for rich event descriptions and context.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Complex Title Examples</h4>
      <div className="space-y-0 border rounded-lg p-4">
        <TimelineEvent
          title={
            <span>
              Deployed <strong>v2.1.0</strong> with{' '}
              <code className="bg-gray-100 dark:bg-gray-800 px-1 rounded text-xs">
                feature-flags
              </code>
            </span>
          }
          status="success"
          createdAt="2024-07-15T12:00:00Z"
          caption="Deployment included new feature flag system"
          createdBy="devops-team"
        />
        
        <TimelineEvent
          title={
            <span>
              <Badge theme="error" size="sm">CRITICAL</Badge>{' '}
              Database connection failed
            </span>
          }
          status="failed"
          createdAt="2024-07-15T11:00:00Z"
          caption="Primary database became unresponsive"
          createdBy="monitoring-alert"
        />
        
        <TimelineEvent
          title={
            <span>
              User <Text family="mono" className="inline">john.doe</Text> updated permissions
            </span>
          }
          status="success"
          createdAt="2024-07-15T10:00:00Z"
          caption="Administrative permissions modified for security role"
          createdBy="admin-panel"
        />
      </div>
    </div>
  </div>
)

export const MinimalConfiguration = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Minimal Configuration</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Timeline events work with minimal required props. Only title,
        status, and createdAt are required - all other props are optional
        and can be omitted for simpler use cases.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Minimal Event</h4>
      <div className="border rounded-lg p-4">
        <TimelineEvent
          title="System Alert"
          status="warn"
          createdAt="2024-07-15T12:00:00Z"
        />
      </div>
      <Text variant="subtext" theme="neutral">
        Only required props: title, status, and createdAt
      </Text>
    </div>
  </div>
)

export const UsageExamples = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Common Usage Patterns</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        TimelineEvent components are typically used within Timeline
        containers to create chronological sequences. They're perfect
        for deployment histories, audit logs, and activity feeds.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Complete Deployment Pipeline</h4>
      <div className="border rounded-lg p-4">
        <div className="space-y-0">
          <TimelineEvent
            title="Production Deployment Complete"
            status="success"
            createdAt="2024-07-15T12:00:00Z"
            caption="Application successfully deployed to production environment"
            additionalCaption="v2.1.0"
            createdBy="deploy-bot"
            badge={{ children: 'Live', theme: 'success' }}
          />
          <TimelineEvent
            title="Security Scan Passed"
            status="success"
            createdAt="2024-07-15T11:45:00Z"
            caption="No vulnerabilities detected in application dependencies"
            createdBy="security-scanner"
          />
          <TimelineEvent
            title="Build Failed - Retrying"
            status="failed"
            createdAt="2024-07-15T11:30:00Z"
            caption="Initial build failed due to network timeout"
            additionalCaption="build-789"
            createdBy="ci-system"
            badge={{ children: 'Auto-Retry', theme: 'warn' }}
          />
          <TimelineEvent
            title="Build Started"
            status="in-progress"
            createdAt="2024-07-15T11:00:00Z"
            caption="Starting automated build process for staging"
            additionalCaption="build-788"
            createdBy="webhook"
          />
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Implementation Example</h4>
      <div className="p-4 border rounded-lg bg-gray-50 dark:bg-gray-800">
        <Text variant="label" theme="neutral" className="mb-2">TypeScript Example:</Text>
        <div className="font-mono text-sm space-y-1">
          <div>{'<TimelineEvent'}</div>
          <div>{'  title="Event Title"'}</div>
          <div>{'  status="success"'}</div>
          <div>{'  createdAt={timestamp}'}</div>
          <div>{'  caption="Event description"'}</div>
          <div>{'  createdBy={user}'}</div>
          <div>{'  additionalCaption={version}'}</div>
          <div>{'  badge={{ children: "Badge" }}'}</div>
          <div>{'/>'}</div>
        </div>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Best Practices:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Use descriptive titles that clearly communicate what happened</li>
        <li>Choose appropriate status types that match the event outcome</li>
        <li>Include creator information for accountability and context</li>
        <li>Use additional captions for version numbers, IDs, or other metadata</li>
        <li>Apply badges consistently for categorization or priority</li>
        <li>Keep descriptions concise but informative</li>
        <li>Ensure timestamps are accurate and properly formatted</li>
      </ul>
    </div>
  </div>
)
