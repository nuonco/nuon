/* eslint-disable react/no-unescaped-entities */
import { CloudPlatform } from './CloudPlatform'
import { Text } from './Text'
import { Badge } from './Badge'

export const BasicUsage = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Basic CloudPlatform Usage</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        CloudPlatform components display cloud provider information with
        consistent branding, icons, and typography. They support all major
        cloud providers and include fallback handling for unknown platforms.
        The component automatically handles icon sizing and text alignment.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">All Supported Platforms</h4>
      <div className="flex items-center gap-6 flex-wrap">
        <div className="text-center space-y-2">
          <CloudPlatform platform="aws" />
          <Text variant="label" className="text-xs">AWS</Text>
        </div>
        <div className="text-center space-y-2">
          <CloudPlatform platform="azure" />
          <Text variant="label" className="text-xs">Azure</Text>
        </div>
        <div className="text-center space-y-2">
          <CloudPlatform platform="gcp" />
          <Text variant="label" className="text-xs">GCP</Text>
        </div>
        <div className="text-center space-y-2">
          <CloudPlatform platform="unknown" />
          <Text variant="label" className="text-xs">Unknown</Text>
        </div>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Platform Support:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>AWS (Amazon Web Services) with official AWS icon</li>
        <li>Azure (Microsoft Azure) with official Azure icon</li>
        <li>GCP (Google Cloud Platform) with official GCP icon</li>
        <li>Unknown platforms with fallback question mark icon</li>
      </ul>
    </div>
  </div>
)

export const DisplayVariants = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">CloudPlatform Display Variants</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          displayVariant
        </code>{' '}
        prop controls how the cloud platform is displayed. Choose between
        abbreviations, full names, or icon-only display based on available
        space and context requirements.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Abbreviation (Default)</h4>
      <div className="flex items-center gap-6 p-4 border rounded">
        <CloudPlatform platform="aws" displayVariant="abbr" />
        <CloudPlatform platform="azure" displayVariant="abbr" />
        <CloudPlatform platform="gcp" displayVariant="abbr" />
        <CloudPlatform platform="unknown" displayVariant="abbr" />
      </div>
      <Text variant="subtext" theme="neutral">
        Compact format using standard abbreviations (AWS, Azure, GCP)
      </Text>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Full Name</h4>
      <div className="flex items-center gap-6 p-4 border rounded flex-wrap">
        <CloudPlatform platform="aws" displayVariant="name" />
        <CloudPlatform platform="azure" displayVariant="name" />
        <CloudPlatform platform="gcp" displayVariant="name" />
        <CloudPlatform platform="unknown" displayVariant="name" />
      </div>
      <Text variant="subtext" theme="neutral">
        Full provider names for clarity and formal contexts
      </Text>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Icon Only</h4>
      <div className="flex items-center gap-6 p-4 border rounded">
        <CloudPlatform platform="aws" displayVariant="icon-only" />
        <CloudPlatform platform="azure" displayVariant="icon-only" />
        <CloudPlatform platform="gcp" displayVariant="icon-only" />
        <CloudPlatform platform="unknown" displayVariant="icon-only" />
      </div>
      <Text variant="subtext" theme="neutral">
        Minimal display with tooltips showing full names on hover
      </Text>
    </div>

    <div className="grid grid-cols-1 md:grid-cols-3 gap-3 text-sm mt-6">
      <div>
        <strong>abbr:</strong> Balanced approach with icons and short text
        (default)
      </div>
      <div>
        <strong>name:</strong> Most descriptive but requires more horizontal
        space
      </div>
      <div>
        <strong>icon-only:</strong> Most compact, ideal for dense interfaces
        and tables
      </div>
    </div>
  </div>
)

export const TextVariants = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">CloudPlatform Text Variants</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        CloudPlatform inherits all Text component props, allowing you to
        customize typography using{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          variant
        </code>
        ,{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          weight
        </code>
        , and{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          theme
        </code>{' '}
        while maintaining consistent cloud platform styling.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Different Text Variants</h4>
      <div className="space-y-3">
        <div className="flex items-center gap-3">
          <CloudPlatform platform="aws" variant="base" />
          <Text variant="subtext">Base variant (default)</Text>
        </div>
        <div className="flex items-center gap-3">
          <CloudPlatform platform="aws" variant="subtext" />
          <Text variant="subtext">Subtext variant</Text>
        </div>
        <div className="flex items-center gap-3">
          <CloudPlatform platform="aws" variant="label" />
          <Text variant="subtext">Label variant</Text>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Different Weight Options</h4>
      <div className="space-y-3">
        <div className="flex items-center gap-3">
          <CloudPlatform platform="gcp" weight="normal" />
          <Text variant="subtext">Normal weight</Text>
        </div>
        <div className="flex items-center gap-3">
          <CloudPlatform platform="gcp" weight="strong" />
          <Text variant="subtext">Strong weight (default)</Text>
        </div>
        <div className="flex items-center gap-3">
          <CloudPlatform platform="gcp" weight="stronger" />
          <Text variant="subtext">Stronger weight</Text>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Different Theme Options</h4>
      <div className="space-y-3">
        <div className="flex items-center gap-3">
          <CloudPlatform platform="azure" theme="brand" />
          <Text variant="subtext">Primary theme</Text>
        </div>
        <div className="flex items-center gap-3">
          <CloudPlatform platform="azure" theme="neutral" />
          <Text variant="subtext">Neutral theme</Text>
        </div>
        <div className="flex items-center gap-3">
          <CloudPlatform platform="azure" theme="neutral" />
          <Text variant="subtext">Muted theme</Text>
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
        CloudPlatform components are commonly used in infrastructure lists,
        deployment interfaces, and configuration settings. Here are typical
        usage patterns for different contexts.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Infrastructure Deployment List</h4>
      <div className="space-y-3">
        <div className="flex items-center justify-between p-3 border rounded">
          <div className="flex items-center gap-3">
            <CloudPlatform platform="aws" />
            <div>
              <Text weight="strong">Production Environment</Text>
              <Text variant="subtext" theme="neutral">us-east-1</Text>
            </div>
          </div>
          <Badge theme="success" size="sm">Active</Badge>
        </div>
        <div className="flex items-center justify-between p-3 border rounded">
          <div className="flex items-center gap-3">
            <CloudPlatform platform="gcp" />
            <div>
              <Text weight="strong">Staging Environment</Text>
              <Text variant="subtext" theme="neutral">us-central1</Text>
            </div>
          </div>
          <Badge theme="warn" size="sm">Updating</Badge>
        </div>
        <div className="flex items-center justify-between p-3 border rounded">
          <div className="flex items-center gap-3">
            <CloudPlatform platform="azure" />
            <div>
              <Text weight="strong">Development Environment</Text>
              <Text variant="subtext" theme="neutral">East US</Text>
            </div>
          </div>
          <Badge theme="info" size="sm">Idle</Badge>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Compact Table View</h4>
      <div className="border rounded overflow-hidden">
        <div className="bg-gray-50 dark:bg-gray-800 px-4 py-2 border-b">
          <div className="grid grid-cols-4 gap-4 text-sm font-medium">
            <Text>Provider</Text>
            <Text>Region</Text>
            <Text>Instance Type</Text>
            <Text>Status</Text>
          </div>
        </div>
        <div className="divide-y">
          <div className="px-4 py-3">
            <div className="grid grid-cols-4 gap-4 items-center text-sm">
              <CloudPlatform platform="aws" displayVariant="icon-only" />
              <Text variant="subtext">us-west-2</Text>
              <Text variant="subtext">t3.medium</Text>
              <Badge theme="success" size="sm">Running</Badge>
            </div>
          </div>
          <div className="px-4 py-3">
            <div className="grid grid-cols-4 gap-4 items-center text-sm">
              <CloudPlatform platform="gcp" displayVariant="icon-only" />
              <Text variant="subtext">europe-west1</Text>
              <Text variant="subtext">n1-standard-2</Text>
              <Badge theme="error" size="sm">Stopped</Badge>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Settings and Configuration</h4>
      <div className="space-y-3 border rounded-lg p-4">
        <div className="flex items-center justify-between">
          <div>
            <Text weight="strong">Default Cloud Provider</Text>
            <Text variant="subtext" theme="neutral">Primary platform for new deployments</Text>
          </div>
          <CloudPlatform platform="aws" displayVariant="name" />
        </div>
        <div className="flex items-center justify-between">
          <div>
            <Text weight="strong">Backup Provider</Text>
            <Text variant="subtext" theme="neutral">Fallback for high availability</Text>
          </div>
          <CloudPlatform platform="gcp" displayVariant="name" />
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Individual Platform Showcases</h4>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div className="p-4 border rounded">
          <Text weight="strong" className="mb-3">Amazon Web Services</Text>
          <div className="flex items-center gap-4">
            <CloudPlatform platform="aws" displayVariant="icon-only" />
            <CloudPlatform platform="aws" displayVariant="abbr" />
            <CloudPlatform platform="aws" displayVariant="name" />
          </div>
        </div>
        <div className="p-4 border rounded">
          <Text weight="strong" className="mb-3">Microsoft Azure</Text>
          <div className="flex items-center gap-4">
            <CloudPlatform platform="azure" displayVariant="icon-only" />
            <CloudPlatform platform="azure" displayVariant="abbr" />
            <CloudPlatform platform="azure" displayVariant="name" />
          </div>
        </div>
        <div className="p-4 border rounded">
          <Text weight="strong" className="mb-3">Google Cloud Platform</Text>
          <div className="flex items-center gap-4">
            <CloudPlatform platform="gcp" displayVariant="icon-only" />
            <CloudPlatform platform="gcp" displayVariant="abbr" />
            <CloudPlatform platform="gcp" displayVariant="name" />
          </div>
        </div>
        <div className="p-4 border rounded">
          <Text weight="strong" className="mb-3">Unknown Platform</Text>
          <div className="flex items-center gap-4">
            <CloudPlatform platform="unknown" displayVariant="icon-only" />
            <CloudPlatform platform="unknown" displayVariant="abbr" />
            <CloudPlatform platform="unknown" displayVariant="name" />
          </div>
        </div>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Best Practices:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Use consistent display variants within the same interface context</li>
        <li>Choose icon-only for tables and compact layouts</li>
        <li>Use full names in settings and configuration interfaces</li>
        <li>Leverage Text component props for consistent typography</li>
        <li>Provide meaningful tooltips for icon-only displays</li>
      </ul>
    </div>
  </div>
)
