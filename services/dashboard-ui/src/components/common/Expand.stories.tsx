/* eslint-disable react/no-unescaped-entities */
import { Expand } from './Expand'
import { Text } from './Text'
import { Icon } from './Icon'
import { Badge } from './Badge'
import { Status } from './Status'

export const BasicUsage = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Basic Expand Usage</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Expand components provide collapsible content sections with smooth
        animations. They feature customizable headings, optional icons, and
        controlled or uncontrolled state management. Perfect for FAQs,
        configuration panels, and detailed information displays.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Simple Expand</h4>
      <div className="max-w-md">
        <Expand
          id="basic-expand"
          heading="Click to expand"
          className="border rounded-lg"
        >
          <div className="p-4">
            <Text>
              This is the expanded content that shows when the expand component is
              opened. The content can include any React elements and will animate
              smoothly when toggled.
            </Text>
          </div>
        </Expand>
      </div>
      <Text variant="subtext" theme="neutral">
        Basic expandable section with default closed state
      </Text>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Key Features:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Smooth expand/collapse animations using CSS transitions</li>
        <li>Keyboard accessibility with proper focus management</li>
        <li>Customizable headings with string or React node content</li>
        <li>Flexible styling with className and headerClassName props</li>
      </ul>
    </div>
  </div>
)

export const StateControl = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">State Control</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          isOpen
        </code>{' '}
        prop controls the initial state of the expand component. This allows
        for programmatic control and default expanded states based on context
        or user preferences.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Initially Closed vs Open</h4>
      <div className="space-y-4 max-w-md">
        <Expand
          id="closed-expand"
          heading="Initially closed (default)"
          className="border rounded-lg"
        >
          <div className="p-4">
            <Text>This expand component starts in a closed state.</Text>
          </div>
        </Expand>
        
        <Expand
          id="open-expand"
          heading="Initially open"
          isOpen={true}
          className="border rounded-lg"
        >
          <div className="p-4">
            <Text>This expand component starts in an open state, showing its content immediately.</Text>
          </div>
        </Expand>
      </div>
    </div>
  </div>
)

export const IconPositioning = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Icon Positioning</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          isIconBeforeHeading
        </code>{' '}
        prop controls whether the expand/collapse caret icon appears before
        or after the heading text. This affects the visual hierarchy and
        scanning pattern for users.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Icon Position Comparison</h4>
      <div className="space-y-3 max-w-md">
        <Expand
          id="icon-after-expand"
          heading="Icon after heading (default)"
          className="border rounded-lg"
        >
          <div className="p-4">
            <Text>The expand/collapse icon appears after the heading text by default.</Text>
          </div>
        </Expand>
        
        <Expand
          id="icon-before-expand"
          heading="Icon before heading"
          isIconBeforeHeading={true}
          className="border rounded-lg"
        >
          <div className="p-4">
            <Text>The expand/collapse icon appears before the heading text when enabled.</Text>
          </div>
        </Expand>
      </div>
    </div>

    <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm mt-6">
      <div>
        <strong>After (default):</strong> Standard pattern with icon on the right,
        consistent with most expand interfaces
      </div>
      <div>
        <strong>Before:</strong> Alternative pattern with icon on the left,
        useful for tree-like structures
      </div>
    </div>
  </div>
)

export const StylingOptions = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Styling Options</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          hasNoHoverStyle
        </code>{' '}
        prop disables hover and focus effects on the header. Use{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          headerClassName
        </code>{' '}
        to customize header styling while maintaining functionality.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Hover Style Comparison</h4>
      <div className="space-y-3 max-w-md">
        <Expand
          id="with-hover-expand"
          heading="With hover effects (default)"
          className="border rounded-lg"
        >
          <div className="p-4">
            <Text>This expand component shows hover and focus effects on the header.</Text>
          </div>
        </Expand>
        
        <Expand
          id="no-hover-expand"
          heading="No hover effects"
          hasNoHoverStyle={true}
          className="border rounded-lg"
        >
          <div className="p-4">
            <Text>This expand component has no hover or focus effects on the header.</Text>
          </div>
        </Expand>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Custom Header Styling</h4>
      <div className="space-y-3 max-w-md">
        <Expand
          id="custom-header-expand"
          heading="Custom styled header"
          headerClassName="bg-blue-50 dark:bg-blue-950 text-blue-900 dark:text-blue-100"
          className="border rounded-lg"
        >
          <div className="p-4">
            <Text>The header has custom background and text colors.</Text>
          </div>
        </Expand>
      </div>
    </div>
  </div>
)

export const CustomContent = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Custom Content</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Expand headings can accept React nodes for rich content including
        icons, badges, status indicators, and complex layouts. This enables
        sophisticated information architecture and visual hierarchy.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Rich Heading Examples</h4>
      <div className="space-y-3 max-w-md">
        <Expand
          id="status-heading-expand"
          className="border rounded-lg"
          heading={
            <div className="flex items-center gap-2">
              <Status status="success" isWithoutText />
              <Text weight="strong">Server Status</Text>
              <Badge theme="success" size="sm">Online</Badge>
            </div>
          }
        >
          <div className="p-4 space-y-2">
            <Text variant="subtext">Last checked: 2 minutes ago</Text>
            <Text variant="subtext">Uptime: 99.9%</Text>
            <Text variant="subtext">Response time: 45ms</Text>
          </div>
        </Expand>
        
        <Expand
          id="warning-heading-expand"
          className="border rounded-lg"
          heading={
            <div className="flex items-center gap-2">
              <Icon variant="Warning" size="16" className="text-orange-600" />
              <Text weight="strong">Configuration Issues</Text>
              <Badge theme="warn" size="sm">3 Issues</Badge>
            </div>
          }
        >
          <div className="p-4 space-y-2">
            <Text variant="subtext">• Environment variable REDIS_URL is missing</Text>
            <Text variant="subtext">• Database connection pool is undersized</Text>
            <Text variant="subtext">• SSL certificate expires in 7 days</Text>
          </div>
        </Expand>
        
        <Expand
          id="user-heading-expand"
          className="border rounded-lg"
          heading={
            <div className="flex items-center justify-between w-full pr-6">
              <div className="flex items-center gap-3">
                <div className="w-8 h-8 bg-gradient-to-br from-blue-500 to-purple-600 rounded-full flex items-center justify-center text-white text-sm font-bold">
                  JD
                </div>
                <div>
                  <Text weight="strong">John Doe</Text>
                  <Text variant="label" theme="neutral">Administrator</Text>
                </div>
              </div>
              <Badge theme="brand" size="sm">Pro</Badge>
            </div>
          }
        >
          <div className="p-4 space-y-3">
            <div className="flex justify-between">
              <Text variant="subtext">Email:</Text>
              <Text variant="subtext">john@example.com</Text>
            </div>
            <div className="flex justify-between">
              <Text variant="subtext">Last login:</Text>
              <Text variant="subtext">2 hours ago</Text>
            </div>
            <div className="flex justify-between">
              <Text variant="subtext">Role:</Text>
              <Text variant="subtext">Administrator</Text>
            </div>
          </div>
        </Expand>
      </div>
    </div>
  </div>
)

export const NestedStructures = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Nested Structures</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Expand components can be nested to create hierarchical information
        structures. This is useful for configuration panels, documentation
        trees, and complex data organization where multiple levels of detail
        are needed.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Configuration Tree</h4>
      <div className="max-w-md">
        <Expand
          id="parent-expand"
          heading="Application Configuration"
          className="border rounded-lg"
          isOpen={true}
        >
          <div className="p-4 space-y-2">
            <Expand id="database-expand" heading="Database Settings" className="border rounded">
              <div className="p-4 space-y-1">
                <div className="flex justify-between">
                  <Text variant="subtext">Host:</Text>
                  <Text variant="subtext">localhost</Text>
                </div>
                <div className="flex justify-between">
                  <Text variant="subtext">Port:</Text>
                  <Text variant="subtext">5432</Text>
                </div>
                <div className="flex justify-between">
                  <Text variant="subtext">Database:</Text>
                  <Text variant="subtext">myapp</Text>
                </div>
              </div>
            </Expand>
            
            <Expand id="api-expand" heading="API Settings" className="border rounded">
              <div className="p-4 space-y-1">
                <div className="flex justify-between">
                  <Text variant="subtext">Base URL:</Text>
                  <Text variant="subtext">https://api.example.com</Text>
                </div>
                <div className="flex justify-between">
                  <Text variant="subtext">Timeout:</Text>
                  <Text variant="subtext">30s</Text>
                </div>
                <div className="flex justify-between">
                  <Text variant="subtext">Rate limit:</Text>
                  <Text variant="subtext">1000/hour</Text>
                </div>
              </div>
            </Expand>
            
            <Expand id="cache-expand" heading="Cache Settings" className="border rounded">
              <div className="p-4 space-y-1">
                <div className="flex justify-between">
                  <Text variant="subtext">Redis URL:</Text>
                  <Text variant="subtext">redis://localhost:6379</Text>
                </div>
                <div className="flex justify-between">
                  <Text variant="subtext">TTL:</Text>
                  <Text variant="subtext">3600s</Text>
                </div>
              </div>
            </Expand>
          </div>
        </Expand>
      </div>
    </div>
  </div>
)

export const UsageExamples = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Common Usage Patterns</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Expand components are versatile and work well in many contexts
        including FAQs, documentation, settings panels, and data visualization.
        Here are common patterns and recommended approaches.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Frequently Asked Questions</h4>
      <div className="max-w-2xl space-y-2">
        <Expand id="faq-1" heading="What is this application?" className="border rounded">
          <div className="p-4">
            <Text>
              This is a modern web application built with Next.js and React that
              helps you manage your projects and workflows efficiently. It provides
              comprehensive tools for deployment, monitoring, and collaboration.
            </Text>
          </div>
        </Expand>
        
        <Expand id="faq-2" heading="How do I get started?" className="border rounded">
          <div className="p-4">
            <Text>
              To get started, create an account, set up your organization, and begin
              by creating your first project. Our onboarding guide will walk you
              through each step, from initial setup to your first deployment.
            </Text>
          </div>
        </Expand>
        
        <Expand id="faq-3" heading="Is my data secure?" className="border rounded">
          <div className="p-4">
            <Text>
              Yes, we take security seriously. All data is encrypted in transit and
              at rest, and we follow industry best practices for data protection and
              privacy. We also provide audit logs and compliance certifications.
            </Text>
          </div>
        </Expand>
        
        <Expand id="faq-4" heading="What are the pricing plans?" className="border rounded">
          <div className="p-4">
            <Text>
              We offer flexible pricing plans including a free tier for small projects,
              professional plans for growing teams, and enterprise solutions for
              large organizations. All plans include core features with different
              usage limits and support levels.
            </Text>
          </div>
        </Expand>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Documentation Sections</h4>
      <div className="max-w-2xl space-y-2">
        <Expand id="docs-1" heading="Getting Started" className="border rounded" isOpen={true}>
          <div className="p-4 space-y-3">
            <Text variant="subtext">Follow these steps to set up your first project:</Text>
            <div className="pl-4 space-y-1">
              <Text variant="subtext">1. Create an account and verify your email</Text>
              <Text variant="subtext">2. Set up your organization</Text>
              <Text variant="subtext">3. Connect your code repository</Text>
              <Text variant="subtext">4. Configure your deployment settings</Text>
              <Text variant="subtext">5. Deploy your first application</Text>
            </div>
          </div>
        </Expand>
        
        <Expand id="docs-2" heading="API Reference" className="border rounded">
          <div className="p-4">
            <Text>
              Comprehensive API documentation with examples, authentication methods,
              rate limits, and response formats. Includes SDKs for popular programming
              languages and interactive examples.
            </Text>
          </div>
        </Expand>
        
        <Expand id="docs-3" heading="Troubleshooting" className="border rounded">
          <div className="p-4">
            <Text>
              Common issues and their solutions, debugging techniques, error code
              references, and support contact information. Updated regularly with
              community-reported issues and resolutions.
            </Text>
          </div>
        </Expand>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Best Practices:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Use clear, descriptive headings that indicate the content within</li>
        <li>Group related content logically and avoid excessive nesting</li>
        <li>Consider initial open/closed states based on importance and context</li>
        <li>Combine with rich content like icons and badges for better UX</li>
        <li>Ensure adequate spacing and visual hierarchy in nested structures</li>
        <li>Test keyboard navigation and screen reader compatibility</li>
      </ul>
    </div>
  </div>
)
