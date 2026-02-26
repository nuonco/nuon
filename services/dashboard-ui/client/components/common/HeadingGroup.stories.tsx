/* eslint-disable react/no-unescaped-entities */
import { HeadingGroup } from './HeadingGroup'
import { Text } from './Text'
import { Badge } from './Badge'
import { Button } from './Button'

export const BasicUsage = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Basic HeadingGroup Usage</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        HeadingGroup components provide semantic grouping for headings and their
        associated descriptive text. Using the HTML hgroup element, they create
        proper document structure and improve accessibility by associating
        related heading content.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Simple Heading Group</h4>
      <div className="p-4 border rounded">
        <HeadingGroup>
          <Text role="heading" level={1} variant="h1" weight="stronger">
            Main Page Title
          </Text>
          <Text variant="subtext" theme="neutral">
            Descriptive subtitle or summary
          </Text>
        </HeadingGroup>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Semantic Benefits:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Uses HTML hgroup element for proper document structure</li>
        <li>Groups related heading content for screen readers</li>
        <li>Flexbox column layout for consistent spacing</li>
        <li>Customizable with standard HTML attributes and CSS classes</li>
      </ul>
    </div>
  </div>
)

export const HeadingHierarchy = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Heading Hierarchy Examples</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        HeadingGroup can contain various heading levels and combinations of
        descriptive text. This helps establish clear information hierarchy and
        improves content organization.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Different Heading Levels</h4>
      <div className="space-y-6">
        <div className="p-4 border rounded">
          <HeadingGroup>
            <Text role="heading" level={1} variant="h1" weight="stronger">
              H1 Main Title
            </Text>
            <Text variant="subtext" theme="neutral">
              Primary page heading with subtitle
            </Text>
          </HeadingGroup>
        </div>

        <div className="p-4 border rounded">
          <HeadingGroup>
            <Text role="heading" level={2} variant="h2" weight="stronger">
              H2 Section Title
            </Text>
            <Text variant="subtext" theme="neutral">
              Section description and context
            </Text>
          </HeadingGroup>
        </div>

        <div className="p-4 border rounded">
          <HeadingGroup>
            <Text role="heading" level={3} variant="h3" weight="stronger">
              H3 Subsection Title
            </Text>
            <Text variant="subtext" theme="neutral">
              Detailed subsection information
            </Text>
          </HeadingGroup>
        </div>
      </div>
    </div>
  </div>
)

export const WithAdditionalContent = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">
        HeadingGroup with Additional Content
      </h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Beyond headings and descriptions, HeadingGroup can contain badges,
        status indicators, and other related elements that provide context or
        actions for the heading section.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">With Badges and Status</h4>
      <div className="space-y-4">
        <div className="p-4 border rounded">
          <HeadingGroup>
            <div className="flex items-center gap-3">
              <Text role="heading" level={2} variant="h2" weight="stronger">
                Project Dashboard
              </Text>
              <Badge theme="success" size="sm">
                Active
              </Badge>
            </div>
            <Text variant="subtext" theme="neutral">
              Overview of project metrics and current status
            </Text>
          </HeadingGroup>
        </div>

        <div className="p-4 border rounded">
          <HeadingGroup>
            <div className="flex items-center gap-3">
              <Text role="heading" level={2} variant="h2" weight="stronger">
                API Documentation
              </Text>
              <Badge theme="brand" size="sm">
                v2.1
              </Badge>
              <Badge theme="info" size="sm">
                Updated
              </Badge>
            </div>
            <Text variant="subtext" theme="neutral">
              Complete API reference and integration guides
            </Text>
            <Text variant="label" theme="neutral">
              Last updated: 2 days ago
            </Text>
          </HeadingGroup>
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
        HeadingGroup is commonly used in page headers, section introductions,
        and content areas where headings need accompanying descriptive text.
        Here are typical patterns for different contexts.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Page Header</h4>
      <div className="p-6 border rounded-lg bg-gray-50 dark:bg-gray-800">
        <HeadingGroup>
          <Text role="heading" level={1} variant="h1" weight="stronger">
            Application Settings
          </Text>
          <Text variant="subtext" theme="neutral">
            Configure your application preferences, integrations, and security
            settings
          </Text>
        </HeadingGroup>
        <div className="mt-6 flex gap-3">
          <Button variant="primary">Save Changes</Button>
          <Button variant="secondary">Reset to Defaults</Button>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Feature Section</h4>
      <div className="p-4 border rounded-lg">
        <HeadingGroup>
          <Text role="heading" level={2} variant="h2" weight="stronger">
            Team Collaboration
          </Text>
          <Text variant="subtext" theme="neutral">
            Work together seamlessly with real-time editing, comments, and
            shared workspaces
          </Text>
        </HeadingGroup>
        <div className="mt-4 space-y-2">
          <Text>• Real-time collaborative editing</Text>
          <Text>• In-line comments and suggestions</Text>
          <Text>• Shared project workspaces</Text>
          <Text>• Team activity tracking</Text>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Article or Blog Post Header</h4>
      <div className="p-4 border rounded-lg">
        <HeadingGroup>
          <Text variant="h1" weight="stronger">
            Getting Started with Modern Web Development
          </Text>
          <Text variant="subtext" theme="neutral">
            A comprehensive guide to building scalable web applications with the
            latest tools and best practices
          </Text>
          <div className="flex items-center gap-4 mt-2">
            <Text variant="label" theme="neutral">
              Published: March 15, 2024
            </Text>
            <Text variant="label" theme="neutral">
              Reading time: 12 min
            </Text>
            <Badge theme="brand" size="sm">
              Featured
            </Badge>
          </div>
        </HeadingGroup>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Dashboard Widget Header</h4>
      <div className="p-4 border rounded-lg">
        <HeadingGroup>
          <div className="flex items-center justify-between">
            <Text variant="h3" weight="stronger">
              Recent Activity
            </Text>
            <Button variant="ghost" size="sm">
              View All
            </Button>
          </div>
          <Text variant="subtext" theme="neutral">
            Latest updates and notifications from your team and projects
          </Text>
        </HeadingGroup>
        <div className="mt-4 space-y-2">
          <Text variant="subtext">
            John updated the design system documentation
          </Text>
          <Text variant="subtext">
            Sarah deployed version 2.1.0 to production
          </Text>
          <Text variant="subtext">Mike created a new API endpoint</Text>
        </div>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Best Practices:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>
          Use for grouping headings with their related descriptive content
        </li>
        <li>Maintain proper heading hierarchy (h1 → h2 → h3)</li>
        <li>Include descriptive text that adds context to the heading</li>
        <li>Consider adding status indicators or metadata when relevant</li>
        <li>Keep grouped content semantically related</li>
      </ul>
    </div>
  </div>
)
