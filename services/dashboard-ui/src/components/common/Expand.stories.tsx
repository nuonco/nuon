import { Expand } from './Expand'
import { Text } from './Text'

export const Default = () => (
  <div className="max-w-md">
    <Expand
      id="basic-expand"
      heading="Click to expand"
      className="border rounded-lg"
    >
      <div className="p-4">
        <Text>
          This is the expanded content that shows when the expand component is
          opened.
        </Text>
      </div>
    </Expand>
  </div>
)

export const InitiallyOpen = () => (
  <div className="max-w-md">
    <Expand
      id="open-expand"
      heading="Initially open"
      isOpen={true}
      className="border rounded-lg"
    >
      <div className="p-4">
        <Text>This expand component starts in an open state.</Text>
      </div>
    </Expand>
  </div>
)

export const IconBeforeHeading = () => (
  <div className="max-w-md">
    <Expand
      id="icon-before-expand"
      heading="Icon first"
      isIconBeforeHeading={true}
      className="border rounded-lg"
    >
      <div className="p-4">
        <Text>The expand/collapse icon appears before the heading text.</Text>
      </div>
    </Expand>
  </div>
)

export const NoHoverEffect = () => (
  <div className="max-w-md">
    <Expand
      id="no-hover-expand"
      heading="No hover effects"
      hasNoHoverStyle={true}
      className="border rounded-lg"
    >
      <div className="p-4">
        <Text>
          This expand component has no hover or focus effects on the header.
        </Text>
      </div>
    </Expand>
  </div>
)

export const CustomHeading = () => (
  <div className="max-w-md">
    <Expand
      id="custom-heading-expand"
      className="border rounded-lg"
      heading={
        <div className="flex items-center gap-2">
          <span className="w-2 h-2 bg-green-500 rounded-full"></span>
          <Text weight="strong">Server Status</Text>
          <span className="text-xs bg-green-100 text-green-800 px-2 py-1 rounded">
            Online
          </span>
        </div>
      }
    >
      <div className="p-4 space-y-2">
        <Text variant="subtext">Last checked: 2 minutes ago</Text>
        <Text variant="subtext">Uptime: 99.9%</Text>
        <Text variant="subtext">Response time: 45ms</Text>
      </div>
    </Expand>
  </div>
)

export const NestedExpands = () => (
  <div className="max-w-md">
    <Expand
      id="parent-expand"
      heading="Configuration"
      className="border rounded-lg"
    >
      <div className="p-4 space-y-2">
        <Expand id="database-expand" heading="Database Settings">
          <div className="p-4">
            <Text variant="subtext">Host: localhost</Text>
            <Text variant="subtext">Port: 5432</Text>
            <Text variant="subtext">Database: myapp</Text>
          </div>
        </Expand>
        <Expand id="api-expand" heading="API Settings">
          <div className="p-4">
            <Text variant="subtext">Base URL: https://api.example.com</Text>
            <Text variant="subtext">Timeout: 30s</Text>
            <Text variant="subtext">Rate limit: 1000/hour</Text>
          </div>
        </Expand>
      </div>
    </Expand>
  </div>
)

export const FAQ = () => (
  <div className="max-w-2xl space-y-2">
    <Expand id="faq-1" heading="What is this application?">
      <div className="p-4">
        <Text>
          This is a modern web application built with Next.js and React that
          helps you manage your projects and workflows efficiently.
        </Text>
      </div>
    </Expand>
    <Expand id="faq-2" heading="How do I get started?">
      <div className="p-4">
        <Text>
          To get started, create an account, set up your organization, and begin
          by creating your first project. Our onboarding guide will walk you
          through each step.
        </Text>
      </div>
    </Expand>
    <Expand id="faq-3" heading="Is my data secure?">
      <div className="p-4">
        <Text>
          Yes, we take security seriously. All data is encrypted in transit and
          at rest, and we follow industry best practices for data protection and
          privacy.
        </Text>
      </div>
    </Expand>
  </div>
)
