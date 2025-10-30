/* eslint-disable react/no-unescaped-entities */
import { StatusWithDescription } from './StatusWithDescription'
import { Text } from './Text'
import { Badge } from './Badge'
import { Button } from './Button'

export const BasicUsage = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Basic StatusWithDescription Usage</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        StatusWithDescription combines a Status component with contextual
        tooltip information. It supports all status variants (dot, badge,
        timeline) and provides hover tooltips for additional details and
        explanations about the status state.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Basic Examples</h4>
      <div className="flex flex-col gap-4 max-w-md">
        <div className="flex items-center gap-3">
          <StatusWithDescription
            statusProps={{ status: 'success' }}
            tooltipProps={{ tipContent: 'Operation completed successfully' }}
          />
          <Text variant="subtext">Deployment successful</Text>
        </div>
        <div className="flex items-center gap-3">
          <StatusWithDescription
            statusProps={{ status: 'error', variant: 'badge' }}
            tooltipProps={{ tipContent: 'Failed to process request - check logs for details' }}
          />
          <Text variant="subtext">Build failed</Text>
        </div>
        <div className="flex items-center gap-3">
          <StatusWithDescription
            statusProps={{ status: 'warn', variant: 'timeline' }}
            tooltipProps={{
              tipContent: 'Warning: Check configuration settings before proceeding',
              position: 'top',
            }}
          />
          <Text variant="subtext">Configuration warning</Text>
        </div>
      </div>
      <Text variant="subtext" theme="neutral">
        Hover over the status indicators to see tooltip information
      </Text>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Key Features:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Combines visual status indicators with descriptive tooltips</li>
        <li>Supports all Status component variants (dot, badge, timeline)</li>
        <li>Configurable tooltip positioning (top, bottom, left, right)</li>
        <li>Full range of status states and color themes</li>
        <li>Accessible hover and focus interactions</li>
      </ul>
    </div>
  </div>
)

export const StatusVariants = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Status Variants and Types</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        StatusWithDescription supports all Status component variants with
        different visual presentations. Each variant serves different use
        cases - dots for compact displays, badges for prominent status,
        and timeline variants for process flows.
      </p>
    </div>

    <div className="space-y-6">
      <div className="space-y-3">
        <h4 className="text-sm font-medium">Dot Variants (Default)</h4>
        <div className="flex items-center gap-6">
          <StatusWithDescription
            statusProps={{ status: 'default' }}
            tooltipProps={{ tipContent: 'Default state - no specific status' }}
          />
          <StatusWithDescription
            statusProps={{ status: 'success' }}
            tooltipProps={{ tipContent: 'Success - operation completed' }}
          />
          <StatusWithDescription
            statusProps={{ status: 'error' }}
            tooltipProps={{ tipContent: 'Error - operation failed' }}
          />
          <StatusWithDescription
            statusProps={{ status: 'warn' }}
            tooltipProps={{ tipContent: 'Warning - needs attention' }}
          />
          <StatusWithDescription
            statusProps={{ status: 'info' }}
            tooltipProps={{ tipContent: 'Info - additional information' }}
          />
          <StatusWithDescription
            statusProps={{ status: 'brand' }}
            tooltipProps={{ tipContent: 'Brand - platform specific status' }}
          />
        </div>
        <Text variant="subtext" theme="neutral">
          Dot variants are compact and work well in lists and tables
        </Text>
      </div>

      <div className="space-y-3">
        <h4 className="text-sm font-medium">Badge Variants</h4>
        <div className="flex items-center gap-4">
          <StatusWithDescription
            statusProps={{ status: 'default', variant: 'badge' }}
            tooltipProps={{ tipContent: 'Default badge status' }}
          />
          <StatusWithDescription
            statusProps={{ status: 'active', variant: 'badge' }}
            tooltipProps={{ tipContent: 'Active status - currently running' }}
          />
          <StatusWithDescription
            statusProps={{ status: 'error', variant: 'badge' }}
            tooltipProps={{ tipContent: 'Error badge - critical failure' }}
          />
          <StatusWithDescription
            statusProps={{ status: 'warn', variant: 'badge' }}
            tooltipProps={{ tipContent: 'Warning badge - requires review' }}
          />
          <StatusWithDescription
            statusProps={{ status: 'info', variant: 'badge' }}
            tooltipProps={{ tipContent: 'Info badge - informational status' }}
          />
          <StatusWithDescription
            statusProps={{ status: 'brand', variant: 'badge' }}
            tooltipProps={{ tipContent: 'Brand badge - Nuon platform status' }}
          />
        </div>
        <Text variant="subtext" theme="neutral">
          Badge variants provide more prominent status display
        </Text>
      </div>

      <div className="space-y-3">
        <h4 className="text-sm font-medium">Timeline Variants</h4>
        <div className="flex items-center gap-6">
          <StatusWithDescription
            statusProps={{ status: 'default', variant: 'timeline' }}
            tooltipProps={{ tipContent: 'Default timeline step' }}
          />
          <StatusWithDescription
            statusProps={{ status: 'success', variant: 'timeline' }}
            tooltipProps={{ tipContent: 'Completed timeline step' }}
          />
          <StatusWithDescription
            statusProps={{ status: 'error', variant: 'timeline' }}
            tooltipProps={{ tipContent: 'Failed timeline step' }}
          />
          <StatusWithDescription
            statusProps={{ status: 'warn', variant: 'timeline' }}
            tooltipProps={{ tipContent: 'Timeline step with warnings' }}
          />
          <StatusWithDescription
            statusProps={{ status: 'info', variant: 'timeline' }}
            tooltipProps={{ tipContent: 'Informational timeline step' }}
          />
          <StatusWithDescription
            statusProps={{ status: 'special', variant: 'timeline' }}
            tooltipProps={{ tipContent: 'Special timeline status' }}
          />
        </div>
        <Text variant="subtext" theme="neutral">
          Timeline variants are designed for process flows and workflows
        </Text>
      </div>
    </div>
  </div>
)

export const TooltipPositions = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Tooltip Positioning</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Tooltips can be positioned on any side of the status indicator
        using the position prop. Choose the appropriate position based
        on the layout and available space around the component.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Tooltip Position Examples</h4>
      <div className="flex items-center justify-center gap-12 p-8 border rounded-lg">
        <div className="text-center space-y-2">
          <StatusWithDescription
            statusProps={{ status: 'success', variant: 'badge' }}
            tooltipProps={{ tipContent: 'Tooltip appears above the status indicator', position: 'top' }}
          />
          <Text variant="subtext">Top</Text>
        </div>
        <div className="text-center space-y-2">
          <StatusWithDescription
            statusProps={{ status: 'error', variant: 'badge' }}
            tooltipProps={{ tipContent: 'Tooltip appears below the status indicator', position: 'bottom' }}
          />
          <Text variant="subtext">Bottom</Text>
        </div>
        <div className="text-center space-y-2">
          <StatusWithDescription
            statusProps={{ status: 'warn', variant: 'badge' }}
            tooltipProps={{ tipContent: 'Tooltip appears to the left of the status indicator', position: 'left' }}
          />
          <Text variant="subtext">Left</Text>
        </div>
        <div className="text-center space-y-2">
          <StatusWithDescription
            statusProps={{ status: 'info', variant: 'badge' }}
            tooltipProps={{ tipContent: 'Tooltip appears to the right of the status indicator', position: 'right' }}
          />
          <Text variant="subtext">Right</Text>
        </div>
      </div>
      <Text variant="subtext" theme="neutral">
        Hover over each status to see the tooltip in different positions
      </Text>
    </div>
  </div>
)

export const DetailedTooltips = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Detailed Tooltip Content</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Tooltips can contain detailed information including timestamps,
        error messages, next steps, and contextual help. This makes
        them perfect for providing actionable information about status states.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Rich Tooltip Examples</h4>
      <div className="space-y-4">
        <div className="flex items-center gap-3">
          <StatusWithDescription
            statusProps={{ status: 'success', statusText: 'Deployed' }}
            tooltipProps={{
              tipContent: 'Successfully deployed v2.1.0 to production at 2:45 PM. All health checks passed.',
              position: 'right'
            }}
          />
          <Text>Production Deployment</Text>
        </div>
        
        <div className="flex items-center gap-3">
          <StatusWithDescription
            statusProps={{ status: 'error', variant: 'badge', statusText: 'Failed' }}
            tooltipProps={{
              tipContent: 'Build failed at step 3/5: TypeScript compilation errors. Check build logs for details.',
              position: 'right'
            }}
          />
          <Text>Build Process</Text>
        </div>
        
        <div className="flex items-center gap-3">
          <StatusWithDescription
            statusProps={{ status: 'warn', variant: 'timeline' }}
            tooltipProps={{
              tipContent: 'Warning: API response time is 2.3s (threshold: 1.5s). Consider optimizing database queries.',
              position: 'right'
            }}
          />
          <Text>API Performance</Text>
        </div>
        
        <div className="flex items-center gap-3">
          <StatusWithDescription
            statusProps={{ status: 'processing', statusText: 'Running' }}
            tooltipProps={{
              tipContent: 'Background job started 5 minutes ago. Estimated completion: 10 minutes remaining.',
              position: 'right'
            }}
          />
          <Text>Background Job</Text>
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
        StatusWithDescription is commonly used in dashboards, monitoring
        interfaces, process flows, and anywhere status information needs
        additional context through tooltips.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Service Health Dashboard</h4>
      <div className="p-4 border rounded-lg space-y-3">
        <Text weight="stronger" className="mb-3">System Status</Text>
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <Text>API Gateway</Text>
            <StatusWithDescription
              statusProps={{ status: 'success' }}
              tooltipProps={{
                tipContent: 'All endpoints healthy. Average response time: 45ms. Uptime: 99.9%',
                position: 'left'
              }}
            />
          </div>
          <div className="flex items-center justify-between">
            <Text>Database</Text>
            <StatusWithDescription
              statusProps={{ status: 'warn' }}
              tooltipProps={{
                tipContent: 'Connection pool at 85% capacity. Consider scaling to handle increased load.',
                position: 'left'
              }}
            />
          </div>
          <div className="flex items-center justify-between">
            <Text>CDN</Text>
            <StatusWithDescription
              statusProps={{ status: 'error' }}
              tooltipProps={{
                tipContent: 'Edge servers in EU region experiencing high latency. Incident #1234 in progress.',
                position: 'left'
              }}
            />
          </div>
          <div className="flex items-center justify-between">
            <Text>Background Jobs</Text>
            <StatusWithDescription
              statusProps={{ status: 'processing', statusText: '4 Running' }}
              tooltipProps={{
                tipContent: '4 jobs running, 12 queued. Oldest job: 2 minutes. No failed jobs in last 24h.',
                position: 'left'
              }}
            />
          </div>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Deployment Pipeline</h4>
      <div className="p-4 border rounded-lg space-y-4">
        <div className="flex justify-between items-center">
          <Text weight="stronger">Release v2.4.1 Pipeline</Text>
          <Button variant="ghost" size="sm">View Logs</Button>
        </div>
        <div className="space-y-4">
          <div className="flex items-center gap-4">
            <StatusWithDescription
              statusProps={{ status: 'success', variant: 'timeline' }}
              tooltipProps={{
                tipContent: 'Build completed successfully in 2m 34s. All dependencies resolved.',
                position: 'right'
              }}
            />
            <div>
              <Text weight="strong">Build</Text>
              <Text variant="subtext" theme="neutral">Completed 5 minutes ago</Text>
            </div>
          </div>
          
          <div className="flex items-center gap-4">
            <StatusWithDescription
              statusProps={{ status: 'success', variant: 'timeline' }}
              tooltipProps={{
                tipContent: 'All 156 unit tests passed. Code coverage: 94%. No security vulnerabilities found.',
                position: 'right'
              }}
            />
            <div>
              <Text weight="strong">Tests</Text>
              <Text variant="subtext" theme="neutral">156/156 passed</Text>
            </div>
          </div>
          
          <div className="flex items-center gap-4">
            <StatusWithDescription
              statusProps={{ status: 'processing', variant: 'timeline' }}
              tooltipProps={{
                tipContent: 'Integration tests in progress. Completed 12/18 test suites. ETA: 3 minutes.',
                position: 'right'
              }}
            />
            <div>
              <Text weight="strong">Integration Tests</Text>
              <Text variant="subtext" theme="neutral">Running... 12/18 complete</Text>
            </div>
          </div>
          
          <div className="flex items-center gap-4">
            <StatusWithDescription
              statusProps={{ status: 'queued', variant: 'timeline' }}
              tooltipProps={{
                tipContent: 'Security scan will start after integration tests complete. Estimated wait: 4 minutes.',
                position: 'right'
              }}
            />
            <div>
              <Text weight="strong">Security Scan</Text>
              <Text variant="subtext" theme="neutral">Queued</Text>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Best Practices:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Provide meaningful tooltip content that adds value beyond the visual status</li>
        <li>Include actionable information like timestamps, next steps, or error details</li>
        <li>Choose appropriate tooltip positions based on layout constraints</li>
        <li>Use consistent status types across similar contexts</li>
        <li>Consider status text for additional quick context</li>
        <li>Keep tooltip content concise but informative</li>
      </ul>
    </div>
  </div>
)
