/* eslint-disable react/no-unescaped-entities */
import { Banner } from './Banner'
import { Button } from './Button'
import { Text } from './Text'

export const AllThemes = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Banner Themes</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">theme</code> prop controls the color scheme and icon of the banner. Each theme includes appropriate icons, colors, and dark mode styling while maintaining accessibility contrast ratios.
      </p>
    </div>
    
    <div className="space-y-4">
      <Banner theme="brand">
        Brand: Welcome to the Nuon platform experience.
      </Banner>
      <Banner theme="error">
        Error: Something went wrong. Please try again.
      </Banner>
      <Banner theme="warn">
        Warning: This action cannot be undone.
      </Banner>
      <Banner theme="info">
        Info: Your changes have been saved automatically.
      </Banner>
      <Banner theme="success">
        Success: Your deployment was completed successfully!
      </Banner>
      <Banner theme="neutral">
        Neutral: System information and general notifications.
      </Banner>
      <Banner theme="default">
        Default: This is a default banner with important information.
      </Banner>
      
      <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm mt-6">
        <div><strong>brand:</strong> Purple primary colors with WarningCircle icon for Nuon platform branding</div>
        <div><strong>error:</strong> Red colors with WarningOctagon icon for error states and critical issues</div>
        <div><strong>warn:</strong> Orange colors with Warning icon for warnings and cautions</div>
        <div><strong>info:</strong> Blue colors with Info icon for informational content</div>
        <div><strong>success:</strong> Green colors with CheckCircle icon for successful operations</div>
        <div><strong>neutral:</strong> Cool grey colors with Info icon for neutral information</div>
        <div><strong>default:</strong> Standard grey colors with Info icon (default if no theme specified)</div>
      </div>
    </div>
  </div>
)

export const ComplexContent = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Complex Banner Content</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Banners support rich content including multiple lines of text, structured information, and action buttons. 
        Use <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">Text</code> components with different variants and 
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">Button</code> components for interactive elements.
      </p>
      <div className="grid grid-cols-1 md:grid-cols-3 gap-3 text-sm">
        <div><strong>Layout:</strong> Use flex layouts with proper spacing for structured content</div>
        <div><strong>Typography:</strong> Mix Text weights and variants for hierarchy</div>
        <div><strong>Actions:</strong> Include relevant buttons with appropriate variants</div>
      </div>
    </div>

    <div className="space-y-4">
      <Banner theme="warn">
        <div className="flex items-center justify-between gap-4">
          <div className="flex flex-col space-y-2">
            <Text weight="strong">Component Approval Required</Text>
            <Text variant="subtext">
              The following components have been updated and require approval before deployment:
            </Text>
            <div className="text-sm space-y-1">
              <div>• API Gateway (v2.1.4) - Security patches applied</div>
              <div>• Database Migration (v1.3.2) - Schema changes detected</div>
              <div>• Frontend Assets (v3.0.1) - UI improvements and bug fixes</div>
            </div>
            <Text variant="subtext" className="text-xs">
              Review each component carefully as this deployment will affect production services.
            </Text>
          </div>
          <div className="flex flex-col gap-2">
            <Button variant="danger" size="sm">Reject All</Button>
            <Button variant="primary" size="sm">Review & Approve</Button>
            <Button variant="secondary" size="sm">View Details</Button>
          </div>
        </div>
      </Banner>

    <Banner theme="error">
      <div className="flex items-start justify-between gap-4">
        <div className="flex flex-col space-y-2">
          <Text weight="strong">Critical Service Outage</Text>
          <Text variant="subtext">
            Multiple services are experiencing issues in the us-east-1 region:
          </Text>
          <div className="text-sm space-y-1">
            <div>🔴 Authentication Service - Down since 14:30 UTC</div>
            <div>🟡 API Gateway - Degraded performance (2x response time)</div>
            <div>🟡 Database Cluster - High latency detected</div>
          </div>
          <Text variant="subtext" className="text-xs">
            Our engineering team is actively investigating. Estimated resolution: 30 minutes.
          </Text>
          <div className="flex items-center gap-2 text-xs">
            <Text variant="subtext">Last updated: 2 minutes ago</Text>
            <Button size="sm">Auto-refresh: ON</Button>
          </div>
        </div>
        <div className="flex gap-2">
          <Button variant="danger" size="sm">View Status Page</Button>
          <Button variant="secondary" size="sm">Subscribe to Updates</Button>
        </div>
      </div>
    </Banner>

    <Banner theme="success">
      <div className="flex items-start justify-between gap-4">
        <div className="flex flex-col space-y-2">
          <Text weight="strong">Deployment Completed Successfully</Text>
          <Text variant="subtext">
            Version 2.1.4 has been deployed to production with the following updates:
          </Text>
          <div className="text-sm space-y-1">
            <div>✅ 12 components updated successfully</div>
            <div>✅ Database migrations applied (0 errors)</div>
            <div>✅ Health checks passed on all services</div>
            <div>✅ SSL certificates renewed automatically</div>
          </div>
          <Text variant="subtext" className="text-xs">
            Total deployment time: 4m 32s | Zero downtime maintained
          </Text>
          <div className="flex items-center gap-4 text-xs">
            <Text variant="subtext">Deployed at: Oct 9, 2024 15:45 UTC</Text>
            <Text variant="subtext">Build: #1247</Text>
          </div>
        </div>
        <div className="flex gap-2">
          <Button variant="primary" size="sm">View Application</Button>
          <Button variant="secondary" size="sm">View Logs</Button>
          <Button variant="ghost" size="sm">Share Success</Button>
        </div>
      </div>
    </Banner>

    <Banner theme="info">
      <div className="flex items-start justify-between gap-4">
        <div className="flex flex-col space-y-2">
          <Text weight="strong">Scheduled Maintenance Window</Text>
          <Text variant="subtext">
            We have scheduled maintenance for infrastructure upgrades:
          </Text>
          <div className="text-sm space-y-1">
            <div><strong>When:</strong> Sunday, October 15, 2024 from 2:00 AM - 6:00 AM UTC</div>
            <div><strong>Duration:</strong> Expected 4 hours (may complete earlier)</div>
            <div><strong>Impact:</strong> Brief service interruptions during database upgrades</div>
          </div>
          <Text variant="subtext" className="text-xs">
            What we're doing: Kubernetes cluster upgrade, database performance improvements, 
            and security patches. All data will be safely backed up before maintenance begins.
          </Text>
          <div className="flex items-center gap-2 text-xs">
            <Text variant="subtext">Notification sent to: engineering@company.com</Text>
          </div>
        </div>
        <div className="flex flex-col gap-2">
          <Button variant="primary" size="sm">Add to Calendar</Button>
          <Button variant="secondary" size="sm">View Maintenance Plan</Button>
          <Button variant="secondary" size="sm">Subscribe to Updates</Button>
        </div>
      </div>
    </Banner>
    
    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Best Practices:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Use appropriate themes that match the content urgency and type</li>
        <li>Keep banner content concise but informative</li>
        <li>Include relevant action buttons when users need to take action</li>
        <li>Structure complex content with proper spacing and hierarchy</li>
        <li>Consider the banner's role in the overall page layout and user flow</li>
      </ul>
    </div>
    </div>
  </div>
)
