/* eslint-disable react/no-unescaped-entities */
import { LabeledStatus } from './LabeledStatus'
import { Text } from './Text'
import { Badge } from './Badge'
import { Button } from './Button'

export const BasicUsage = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Basic LabeledStatus Usage</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        LabeledStatus combines a descriptive label with a Status component and
        optional tooltip. It's perfect for displaying system states, process
        statuses, and operational indicators with clear labeling and contextual
        information.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Simple Example</h4>
      <div className="flex items-center gap-6">
        <LabeledStatus
          label="Service Status"
          statusProps={{ status: 'success' }}
          tooltipProps={{ tipContent: 'Service is running normally' }}
        />
        <LabeledStatus
          label="Database"
          statusProps={{ status: 'success' }}
          tooltipProps={{ tipContent: 'Database connection is healthy' }}
        />
      </div>
      <Text variant="subtext" theme="neutral">
        Click the status indicators to see tooltip information
      </Text>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Key Features:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Combines descriptive labels with visual status indicators</li>
        <li>Supports tooltips for additional context and information</li>
        <li>Consistent spacing and alignment for multiple status displays</li>
        <li>
          Full range of status states (success, failed, warn, queued, etc.)
        </li>
      </ul>
    </div>
  </div>
)

export const StatusVariants = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Status Variants</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        LabeledStatus supports all Status component variants with appropriate
        colors and styling. Each status type conveys different operational
        states and conditions.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">All Status Types</h4>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <LabeledStatus
          label="Success State"
          statusProps={{ status: 'success' }}
          tooltipProps={{ tipContent: 'Operation completed successfully' }}
        />
        <LabeledStatus
          label="Failed State"
          statusProps={{ status: 'failed' }}
          tooltipProps={{
            tipContent: 'Operation failed - check logs for details',
          }}
        />
        <LabeledStatus
          label="Warning State"
          statusProps={{ status: 'warn' }}
          tooltipProps={{ tipContent: 'Operation completed with warnings' }}
        />
        <LabeledStatus
          label="Queued State"
          statusProps={{ status: 'queued' }}
          tooltipProps={{ tipContent: 'Operation is queued for execution' }}
        />
        <LabeledStatus
          label="In Progress"
          statusProps={{ status: 'processing' }}
          tooltipProps={{ tipContent: 'Operation is currently running' }}
        />
        <LabeledStatus
          label="Unknown State"
          statusProps={{ status: 'unknown' }}
          tooltipProps={{ tipContent: 'Status could not be determined' }}
        />
      </div>
    </div>
  </div>
)

export const WithStatusText = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Status with Text</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Status components can include custom text alongside the visual
        indicator. This provides additional context directly in the status
        display without requiring tooltip interaction.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Status with Custom Text</h4>
      <div className="space-y-3">
        <LabeledStatus
          label="Deployment Status"
          statusProps={{ status: 'success', statusText: 'Deployed' }}
          tooltipProps={{ tipContent: 'Last deployed 2 minutes ago' }}
        />
        <LabeledStatus
          label="Build Status"
          statusProps={{ status: 'processing', statusText: 'Building...' }}
          tooltipProps={{ tipContent: 'Build started 30 seconds ago' }}
        />
        <LabeledStatus
          label="Test Status"
          statusProps={{ status: 'failed', statusText: '3 Failed' }}
          tooltipProps={{ tipContent: '3 tests failed out of 15 total tests' }}
        />
        <LabeledStatus
          label="Queue Position"
          statusProps={{ status: 'queued', statusText: '#4 in queue' }}
          tooltipProps={{ tipContent: 'Estimated wait time: 2-3 minutes' }}
        />
      </div>
    </div>
  </div>
)

export const SystemMonitoring = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">System Monitoring Example</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        LabeledStatus is commonly used in monitoring dashboards, system health
        displays, and operational interfaces where multiple services or
        components need status visualization.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Service Health Dashboard</h4>
      <div className="p-4 border rounded-lg space-y-4">
        <Text weight="stronger" className="mb-3">
          Infrastructure Status
        </Text>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <LabeledStatus
            label="API Gateway"
            statusProps={{ status: 'success', statusText: 'Healthy' }}
            tooltipProps={{ tipContent: 'Response time: 45ms, Uptime: 99.9%' }}
          />
          <LabeledStatus
            label="Database"
            statusProps={{ status: 'success', statusText: 'Connected' }}
            tooltipProps={{
              tipContent: 'Connections: 12/100, Query time: 2ms avg',
            }}
          />
          <LabeledStatus
            label="Redis Cache"
            statusProps={{ status: 'warn', statusText: 'High Memory' }}
            tooltipProps={{ tipContent: 'Memory usage: 85%, consider scaling' }}
          />
          <LabeledStatus
            label="Background Jobs"
            statusProps={{ status: 'processing', statusText: '4 Running' }}
            tooltipProps={{ tipContent: '4 jobs running, 12 queued' }}
          />
          <LabeledStatus
            label="File Storage"
            statusProps={{ status: 'success', statusText: 'Available' }}
            tooltipProps={{
              tipContent: 'Usage: 42GB/100GB, All regions healthy',
            }}
          />
          <LabeledStatus
            label="CDN"
            statusProps={{ status: 'failed', statusText: 'Degraded' }}
            tooltipProps={{
              tipContent: 'Some edge locations experiencing issues',
            }}
          />
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Deployment Pipeline</h4>
      <div className="p-4 border rounded-lg space-y-4">
        <Text weight="stronger" className="mb-3">
          Release Pipeline - v2.4.1
        </Text>
        <div className="space-y-3">
          <LabeledStatus
            label="Code Build"
            statusProps={{ status: 'success', statusText: 'Passed' }}
            tooltipProps={{ tipContent: 'Build completed in 2m 34s' }}
          />
          <LabeledStatus
            label="Unit Tests"
            statusProps={{ status: 'success', statusText: '156/156' }}
            tooltipProps={{
              tipContent: 'All unit tests passed, Coverage: 94%',
            }}
          />
          <LabeledStatus
            label="Integration Tests"
            statusProps={{ status: 'processing', statusText: 'Running...' }}
            tooltipProps={{
              tipContent: 'Progress: 12/18 test suites completed',
            }}
          />
          <LabeledStatus
            label="Security Scan"
            statusProps={{ status: 'queued', statusText: 'Pending' }}
            tooltipProps={{
              tipContent: 'Waiting for integration tests to complete',
            }}
          />
          <LabeledStatus
            label="Deploy to Staging"
            statusProps={{ status: 'queued', statusText: 'Scheduled' }}
            tooltipProps={{
              tipContent: 'Deployment scheduled after security scan',
            }}
          />
        </div>
      </div>
    </div>
  </div>
)

export const UsageExamples = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Common Usage Patterns</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        LabeledStatus is versatile and can be used in various contexts including
        system monitoring, workflow tracking, resource status, and operational
        dashboards.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Resource Status Panel</h4>
      <div className="p-4 border rounded-lg space-y-3">
        <div className="flex justify-between items-center mb-3">
          <Text weight="stronger">Server Resources</Text>
          <Button variant="ghost" size="sm">
            Refresh
          </Button>
        </div>
        <div className="grid grid-cols-2 gap-3">
          <LabeledStatus
            label="CPU Usage"
            statusProps={{ status: 'success', statusText: '34%' }}
            tooltipProps={{ tipContent: 'CPU usage is within normal range' }}
          />
          <LabeledStatus
            label="Memory"
            statusProps={{ status: 'warn', statusText: '78%' }}
            tooltipProps={{ tipContent: 'Memory usage is approaching limits' }}
          />
          <LabeledStatus
            label="Disk Space"
            statusProps={{ status: 'success', statusText: '45%' }}
            tooltipProps={{ tipContent: 'Plenty of disk space available' }}
          />
          <LabeledStatus
            label="Network"
            statusProps={{ status: 'success', statusText: 'Normal' }}
            tooltipProps={{ tipContent: 'Network connectivity is stable' }}
          />
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Feature Flags</h4>
      <div className="p-4 border rounded-lg space-y-3">
        <Text weight="stronger" className="mb-3">
          Feature Toggle Status
        </Text>
        <div className="space-y-2">
          <LabeledStatus
            label="New Dashboard UI"
            statusProps={{ status: 'success', statusText: 'Enabled' }}
            tooltipProps={{ tipContent: 'Feature is active for 100% of users' }}
          />
          <LabeledStatus
            label="Beta API v3"
            statusProps={{ status: 'warn', statusText: 'Partial' }}
            tooltipProps={{ tipContent: 'Feature is active for 25% of users' }}
          />
          <LabeledStatus
            label="Advanced Analytics"
            statusProps={{ status: 'queued', statusText: 'Scheduled' }}
            tooltipProps={{
              tipContent: 'Feature rollout scheduled for next week',
            }}
          />
          <LabeledStatus
            label="Legacy Mode"
            statusProps={{ status: 'failed', statusText: 'Disabled' }}
            tooltipProps={{
              tipContent: 'Feature has been deprecated and disabled',
            }}
          />
        </div>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Best Practices:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>
          Use descriptive labels that clearly identify what's being monitored
        </li>
        <li>Provide meaningful tooltips with actionable information</li>
        <li>Group related status indicators logically</li>
        <li>Use consistent status types for similar operational states</li>
        <li>Include additional context in status text when helpful</li>
        <li>Consider refresh mechanisms for real-time status updates</li>
      </ul>
    </div>
  </div>
)
