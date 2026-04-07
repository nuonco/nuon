export default {
  title: 'Components/ComponentType',
}

import { ComponentType } from './ComponentType'
import { Text } from '../common/Text'
import { Badge } from '../common/Badge'

export const ColorVariants = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">ComponentType color variants</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          colorVariant
        </code>{' '}
        prop switches between monochrome (default) and full brand-color icons.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Mono (default)</h4>
      <div className="flex items-center gap-6 p-4 border rounded flex-wrap">
        <ComponentType type="docker_build" />
        <ComponentType type="external_image" />
        <ComponentType type="helm_chart" />
        <ComponentType type="terraform_module" />
        <ComponentType type={'pulumi_module' as any} />
        <ComponentType type="job" />
        <ComponentType type="kubernetes_manifest" />
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Color</h4>
      <div className="flex items-center gap-6 p-4 border rounded flex-wrap">
        <ComponentType type="docker_build" colorVariant="color" />
        <ComponentType type="external_image" colorVariant="color" />
        <ComponentType type="helm_chart" colorVariant="color" />
        <ComponentType type="terraform_module" colorVariant="color" />
        <ComponentType type={'pulumi_module' as any} colorVariant="color" />
        <ComponentType type="job" colorVariant="color" />
        <ComponentType type="kubernetes_manifest" colorVariant="color" />
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Color with icon-only</h4>
      <div className="flex items-center gap-6 p-4 border rounded flex-wrap">
        <ComponentType type="docker_build" colorVariant="color" displayVariant="icon-only" />
        <ComponentType type="external_image" colorVariant="color" displayVariant="icon-only" />
        <ComponentType type="helm_chart" colorVariant="color" displayVariant="icon-only" />
        <ComponentType type="terraform_module" colorVariant="color" displayVariant="icon-only" />
        <ComponentType type={'pulumi_module' as any} colorVariant="color" displayVariant="icon-only" />
        <ComponentType type="job" colorVariant="color" displayVariant="icon-only" />
        <ComponentType type="kubernetes_manifest" colorVariant="color" displayVariant="icon-only" />
      </div>
    </div>
  </div>
)

export const BasicUsage = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Basic ComponentType Usage</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        ComponentType components display infrastructure component types with
        consistent icons, labels, and typography. They support all component
        types on the Nuon platform and include fallback handling for unknown
        types. The component automatically handles icon sizing and text
        alignment.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">All Supported Types</h4>
      <div className="flex items-center gap-6 flex-wrap">
        <div className="text-center space-y-2">
          <ComponentType type="docker_build" />
          <Text variant="label" className="text-xs">
            Docker
          </Text>
        </div>
        <div className="text-center space-y-2">
          <ComponentType type="external_image" />
          <Text variant="label" className="text-xs">
            External Image
          </Text>
        </div>
        <div className="text-center space-y-2">
          <ComponentType type="helm_chart" />
          <Text variant="label" className="text-xs">
            Helm
          </Text>
        </div>
        <div className="text-center space-y-2">
          <ComponentType type="terraform_module" />
          <Text variant="label" className="text-xs">
            Terraform
          </Text>
        </div>
        <div className="text-center space-y-2">
          <ComponentType type={'pulumi_module' as any} />
          <Text variant="label" className="text-xs">
            Pulumi
          </Text>
        </div>
        <div className="text-center space-y-2">
          <ComponentType type="job" />
          <Text variant="label" className="text-xs">
            Job
          </Text>
        </div>
        <div className="text-center space-y-2">
          <ComponentType type="kubernetes_manifest" />
          <Text variant="label" className="text-xs">
            Kubernetes
          </Text>
        </div>
        <div className="text-center space-y-2">
          <ComponentType type="unknown" />
          <Text variant="label" className="text-xs">
            Unknown
          </Text>
        </div>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Component Types:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Docker Build — container image built from a Dockerfile</li>
        <li>External Image — pre-built OCI container image</li>
        <li>Helm Chart — Kubernetes Helm chart deployment</li>
        <li>Terraform Module — infrastructure as code module</li>
        <li>Pulumi Module — infrastructure as code with general-purpose languages</li>
        <li>Job — Lambda function execution</li>
        <li>Kubernetes Manifest — raw Kubernetes YAML manifest</li>
        <li>Unknown — fallback for unrecognized component types</li>
      </ul>
    </div>
  </div>
)

export const DisplayVariants = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">ComponentType Display Variants</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          displayVariant
        </code>{' '}
        prop controls how the component type is displayed. Choose between full
        names, abbreviations, or icon-only display based on available space and
        context requirements.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Full Name (Default)</h4>
      <div className="flex items-center gap-6 p-4 border rounded flex-wrap">
        <ComponentType type="docker_build" displayVariant="name" />
        <ComponentType type="external_image" displayVariant="name" />
        <ComponentType type="helm_chart" displayVariant="name" />
        <ComponentType type="terraform_module" displayVariant="name" />
        <ComponentType type={'pulumi_module' as any} displayVariant="name" />
        <ComponentType type="job" displayVariant="name" />
        <ComponentType type="kubernetes_manifest" displayVariant="name" />
      </div>
      <Text variant="subtext" theme="neutral">
        Full descriptive names for clarity and formal contexts
      </Text>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Abbreviation</h4>
      <div className="flex items-center gap-6 p-4 border rounded">
        <ComponentType type="docker_build" displayVariant="abbr" />
        <ComponentType type="external_image" displayVariant="abbr" />
        <ComponentType type="helm_chart" displayVariant="abbr" />
        <ComponentType type="terraform_module" displayVariant="abbr" />
        <ComponentType type={'pulumi_module' as any} displayVariant="abbr" />
        <ComponentType type="job" displayVariant="abbr" />
        <ComponentType type="kubernetes_manifest" displayVariant="abbr" />
      </div>
      <Text variant="subtext" theme="neutral">
        Compact format using standard abbreviations (Docker, OCI, Helm, TF, Job,
        K8s)
      </Text>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Icon Only</h4>
      <div className="flex items-center gap-6 p-4 border rounded">
        <ComponentType type="docker_build" displayVariant="icon-only" />
        <ComponentType type="external_image" displayVariant="icon-only" />
        <ComponentType type="helm_chart" displayVariant="icon-only" />
        <ComponentType type="terraform_module" displayVariant="icon-only" />
        <ComponentType type={'pulumi_module' as any} displayVariant="icon-only" />
        <ComponentType type="job" displayVariant="icon-only" />
        <ComponentType type="kubernetes_manifest" displayVariant="icon-only" />
      </div>
      <Text variant="subtext" theme="neutral">
        Minimal display with tooltips showing full names on hover
      </Text>
    </div>

    <div className="grid grid-cols-1 md:grid-cols-3 gap-3 text-sm mt-6">
      <div>
        <strong>name:</strong> Most descriptive with icon and full label
        (default)
      </div>
      <div>
        <strong>abbr:</strong> Balanced approach with icons and short text
      </div>
      <div>
        <strong>icon-only:</strong> Most compact, ideal for dense interfaces and
        tables
      </div>
    </div>
  </div>
)

export const TextVariants = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">ComponentType Text Variants</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        ComponentType inherits all Text component props, allowing you to
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
        while maintaining consistent component type styling.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Different Text Variants</h4>
      <div className="space-y-3">
        <div className="flex items-center gap-3">
          <ComponentType type="docker_build" variant="base" />
          <Text variant="subtext">Base variant (default)</Text>
        </div>
        <div className="flex items-center gap-3">
          <ComponentType type="docker_build" variant="subtext" />
          <Text variant="subtext">Subtext variant</Text>
        </div>
        <div className="flex items-center gap-3">
          <ComponentType type="docker_build" variant="label" />
          <Text variant="subtext">Label variant</Text>
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
        ComponentType components are commonly used in app configuration,
        component lists, and build interfaces. Here are typical usage patterns
        for different contexts.
      </p>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Component List</h4>
      <div className="space-y-3">
        <div className="flex items-center justify-between p-3 border rounded">
          <div className="flex items-center gap-3">
            <ComponentType type="docker_build" />
            <div>
              <Text weight="strong">api-gateway</Text>
              <Text variant="subtext" theme="neutral">
                Main API service
              </Text>
            </div>
          </div>
          <Badge theme="success" size="sm">
            Active
          </Badge>
        </div>
        <div className="flex items-center justify-between p-3 border rounded">
          <div className="flex items-center gap-3">
            <ComponentType type="helm_chart" />
            <div>
              <Text weight="strong">monitoring-stack</Text>
              <Text variant="subtext" theme="neutral">
                Prometheus + Grafana
              </Text>
            </div>
          </div>
          <Badge theme="warn" size="sm">
            Updating
          </Badge>
        </div>
        <div className="flex items-center justify-between p-3 border rounded">
          <div className="flex items-center gap-3">
            <ComponentType type="terraform_module" />
            <div>
              <Text weight="strong">vpc-network</Text>
              <Text variant="subtext" theme="neutral">
                Core networking infrastructure
              </Text>
            </div>
          </div>
          <Badge theme="info" size="sm">
            Idle
          </Badge>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Compact Table View</h4>
      <div className="border rounded overflow-hidden">
        <div className="bg-gray-50 dark:bg-gray-800 px-4 py-2 border-b">
          <div className="grid grid-cols-4 gap-4 text-sm font-medium">
            <Text>Type</Text>
            <Text>Name</Text>
            <Text>Version</Text>
            <Text>Status</Text>
          </div>
        </div>
        <div className="divide-y">
          <div className="px-4 py-3">
            <div className="grid grid-cols-4 gap-4 items-center text-sm">
              <ComponentType type="docker_build" displayVariant="icon-only" />
              <Text variant="subtext">api-gateway</Text>
              <Text variant="subtext">v2.1.4</Text>
              <Badge theme="success" size="sm">
                Running
              </Badge>
            </div>
          </div>
          <div className="px-4 py-3">
            <div className="grid grid-cols-4 gap-4 items-center text-sm">
              <ComponentType type="kubernetes_manifest" displayVariant="icon-only" />
              <Text variant="subtext">ingress-controller</Text>
              <Text variant="subtext">v1.0.0</Text>
              <Badge theme="error" size="sm">
                Failed
              </Badge>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div className="space-y-4">
      <h4 className="text-sm font-medium">Individual Type Showcases</h4>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div className="p-4 border rounded">
          <Text weight="strong" className="mb-3">
            Docker Build
          </Text>
          <div className="flex items-center gap-4">
            <ComponentType type="docker_build" displayVariant="icon-only" />
            <ComponentType type="docker_build" displayVariant="abbr" />
            <ComponentType type="docker_build" displayVariant="name" />
          </div>
        </div>
        <div className="p-4 border rounded">
          <Text weight="strong" className="mb-3">
            External Image
          </Text>
          <div className="flex items-center gap-4">
            <ComponentType type="external_image" displayVariant="icon-only" />
            <ComponentType type="external_image" displayVariant="abbr" />
            <ComponentType type="external_image" displayVariant="name" />
          </div>
        </div>
        <div className="p-4 border rounded">
          <Text weight="strong" className="mb-3">
            Helm Chart
          </Text>
          <div className="flex items-center gap-4">
            <ComponentType type="helm_chart" displayVariant="icon-only" />
            <ComponentType type="helm_chart" displayVariant="abbr" />
            <ComponentType type="helm_chart" displayVariant="name" />
          </div>
        </div>
        <div className="p-4 border rounded">
          <Text weight="strong" className="mb-3">
            Terraform Module
          </Text>
          <div className="flex items-center gap-4">
            <ComponentType type="terraform_module" displayVariant="icon-only" />
            <ComponentType type="terraform_module" displayVariant="abbr" />
            <ComponentType type="terraform_module" displayVariant="name" />
          </div>
        </div>
        <div className="p-4 border rounded">
          <Text weight="strong" className="mb-3">
            Pulumi Module
          </Text>
          <div className="flex items-center gap-4">
            <ComponentType type={'pulumi_module' as any} displayVariant="icon-only" />
            <ComponentType type={'pulumi_module' as any} displayVariant="abbr" />
            <ComponentType type={'pulumi_module' as any} displayVariant="name" />
          </div>
        </div>
        <div className="p-4 border rounded">
          <Text weight="strong" className="mb-3">
            Job (Lambda)
          </Text>
          <div className="flex items-center gap-4">
            <ComponentType type="job" displayVariant="icon-only" />
            <ComponentType type="job" displayVariant="abbr" />
            <ComponentType type="job" displayVariant="name" />
          </div>
        </div>
        <div className="p-4 border rounded">
          <Text weight="strong" className="mb-3">
            Kubernetes Manifest
          </Text>
          <div className="flex items-center gap-4">
            <ComponentType type="kubernetes_manifest" displayVariant="icon-only" />
            <ComponentType type="kubernetes_manifest" displayVariant="abbr" />
            <ComponentType type="kubernetes_manifest" displayVariant="name" />
          </div>
        </div>
      </div>
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Best Practices:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>
          Use consistent display variants within the same interface context
        </li>
        <li>Choose icon-only for tables and compact layouts</li>
        <li>Use full names in detail views and configuration interfaces</li>
        <li>Leverage Text component props for consistent typography</li>
        <li>Provide meaningful tooltips for icon-only displays</li>
      </ul>
    </div>
  </div>
)
