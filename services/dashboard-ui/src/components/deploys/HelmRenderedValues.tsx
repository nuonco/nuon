'use client'

import React, { useState } from 'react'
import { CodeBlock } from '@/components/common/CodeBlock'

// Types
interface HelmDeploymentData {
  deployments: Record<string, any>
  services: Record<string, any>
  ingresses: Record<string, any>
  resources: Record<string, any>
  manifest: string
}

interface TKeyValueWithType {
  key: string
  value: string
  type: string
}

// Utility Functions
function objectToKeyValueArray(obj: Record<string, any>): TKeyValueWithType[] {
  return Object.entries(obj).map(([key, value]) => ({
    key,
    value: formatValue(value),
    type: getValueType(value),
  }))
}

function getValueType(value: any): string {
  if (value === null) return 'null'
  if (value === undefined) return 'undefined'
  if (Array.isArray(value)) return 'array'
  if (typeof value === 'object') return 'object'
  return typeof value
}

function formatValue(value: any): string {
  if (value === null) return 'null'
  if (value === undefined) return 'undefined'
  if (typeof value === 'string') return value
  if (typeof value === 'boolean' || typeof value === 'number')
    return String(value)

  if (typeof value === 'object') {
    if (Array.isArray(value)) {
      return `[${value.map((item) => formatValue(item)).join(', ')}]`
    }
    try {
      return JSON.stringify(value, null, 2)
    } catch (error) {
      return '[Object - Unable to serialize]'
    }
  }

  if (typeof value === 'function') return '[Function]'
  return String(value)
}

function getDeploymentStatus(deployments: Record<string, any>): string {
  for (const namespace of Object.values(deployments)) {
    for (const deployment of Object.values(namespace as any)) {
      const status = (deployment as any).status
      const replicas = {
        desired: status?.replicas || 0,
        ready: status?.readyReplicas || 0,
        available: status?.availableReplicas || 0,
      }

      if (
        replicas.ready !== replicas.desired ||
        replicas.available !== replicas.desired
      ) {
        return 'pending'
      }
    }
  }
  return 'healthy'
}

function cn(...classes: (string | undefined | false)[]): string {
  return classes.filter(Boolean).join(' ')
}

// Components
function StatusBadge({ status }: { status: string }) {
  const statusConfig = {
    healthy: { bg: 'bg-green-100', text: 'text-green-800' },
    pending: { bg: 'bg-yellow-100', text: 'text-yellow-800' },
    error: { bg: 'bg-red-100', text: 'text-red-800' },
    none: { bg: 'bg-gray-100', text: 'text-gray-600' },
  }

  const config =
    statusConfig[status as keyof typeof statusConfig] || statusConfig.none

  return (
    <span
      className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${config.bg} ${config.text}`}
    >
      {status.toUpperCase()}
    </span>
  )
}

function KeyValueList({ values }: { values: TKeyValueWithType[] }) {
  return values?.length ? (
    <div className="grid grid-cols-[max-content_1fr] gap-0 text-sm">
      {/* Header row */}
      <div className="py-2 border-b font-medium text-gray-600">Name</div>
      <div className="py-2 pl-4 border-b font-medium text-gray-600">Value</div>

      {/* Data rows */}
      {values.map(({ key, value, type }, index) => {
        const isLast = index === values.length - 1
        const isObject = type === 'object' || type === 'array'

        return (
          <React.Fragment key={key}>
            <div
              className={cn(
                'py-2 break-all whitespace-nowrap font-mono text-sm',
                !isLast && 'border-b'
              )}
            >
              {key}
            </div>
            <div
              className={cn(
                'py-2 pl-4 break-all font-mono text-sm',
                !isLast && 'border-b',
                isObject && 'whitespace-pre-wrap'
              )}
            >
              {value || <span className="text-gray-400">—</span>}
            </div>
          </React.Fragment>
        )
      })}
    </div>
  ) : (
    <div className="text-gray-500 text-sm">No data available</div>
  )
}

function ResourceCard({
  title,
  count,
  color,
  status,
}: {
  title: string
  count: number
  color: string
  status: string
}) {
  const colorClasses = {
    blue: 'bg-blue-50 border-blue-200',
    green: 'bg-green-50 border-green-200',
    purple: 'bg-purple-50 border-purple-200',
    gray: 'bg-gray-50 border-gray-200',
  }

  return (
    <div
      className={`rounded-lg border p-4 ${colorClasses[color as keyof typeof colorClasses]}`}
    >
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm font-medium text-gray-600">{title}</p>
          <p className="text-2xl font-bold">{count}</p>
        </div>
      </div>
      <div className="mt-2">
        <StatusBadge status={status} />
      </div>
    </div>
  )
}

function DeploymentHeader({ data }: { data: HelmDeploymentData }) {
  // Extract release info from first deployment
  const firstDeployment = Object.values(
    Object.values(data.deployments)[0] || {}
  )[0] as any
  const releaseName =
    firstDeployment?.metadata?.annotations?.['meta.helm.sh/release-name'] ||
    'Unknown'
  const namespace =
    firstDeployment?.metadata?.annotations?.[
      'meta.helm.sh/release-namespace'
    ] || 'Unknown'

  return (
    <div className="bg-white rounded-lg border p-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Helm Deployment</h1>
          <div className="mt-1 flex items-center space-x-4 text-sm text-gray-500">
            <span>
              Release: <span className="font-medium">{releaseName}</span>
            </span>
            <span>
              Namespace: <span className="font-medium">{namespace}</span>
            </span>
            <span>Generated: job.create_at</span>
          </div>
        </div>
        <div className="text-right">
          <div className="text-sm text-gray-500">Overall Status</div>
          <StatusBadge status={getDeploymentStatus(data.deployments)} />
        </div>
      </div>
    </div>
  )
}

function DeploymentStatusSummary({
  deployments,
}: {
  deployments: Record<string, any>
}) {
  const allDeployments = Object.values(deployments).flatMap((namespace) =>
    Object.entries(namespace as any)
  )

  return (
    <div className="space-y-3">
      {allDeployments.map(([name, deployment]: [string, any]) => {
        const status = deployment.status
        const replicas = {
          desired: status?.replicas || 0,
          ready: status?.readyReplicas || 0,
          available: status?.availableReplicas || 0,
        }

        const isHealthy =
          replicas.ready === replicas.desired &&
          replicas.available === replicas.desired
        const healthPercentage =
          replicas.desired > 0
            ? Math.round((replicas.ready / replicas.desired) * 100)
            : 0

        return (
          <div key={name} className="flex items-center justify-between">
            <div className="flex items-center space-x-3">
              <div
                className={`w-3 h-3 rounded-full ${isHealthy ? 'bg-green-400' : 'bg-yellow-400'}`}
              />
              <span className="font-medium">{name}</span>
              <span className="text-sm text-gray-500">
                {replicas.ready}/{replicas.desired} replicas
              </span>
            </div>
            <div className="flex items-center space-x-3">
              <div className="w-24 bg-gray-200 rounded-full h-2">
                <div
                  className={`h-2 rounded-full transition-all duration-500 ${
                    isHealthy ? 'bg-green-500' : 'bg-yellow-500'
                  }`}
                  style={{ width: `${healthPercentage}%` }}
                />
              </div>
              <span className="text-sm font-medium text-gray-600">
                {healthPercentage}%
              </span>
            </div>
          </div>
        )
      })}
    </div>
  )
}

function OverviewTab({ data }: { data: HelmDeploymentData }) {
  return (
    <div className="space-y-6">
      {/* Resource Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <ResourceCard
          title="Deployments"
          count={Object.keys(data.deployments).length}
          color="blue"
          status={getDeploymentStatus(data.deployments)}
        />
        <ResourceCard
          title="Services"
          count={Object.keys(data.services).length}
          color="green"
          status="healthy"
        />
        <ResourceCard
          title="Ingresses"
          count={Object.keys(data.ingresses).length}
          color="purple"
          status="none"
        />
        <ResourceCard
          title="Resources"
          count={Object.keys(data.resources).length}
          color="gray"
          status="healthy"
        />
      </div>

      {/* Quick Status */}
      <div className="bg-white rounded-lg border p-6">
        <h3 className="text-lg font-semibold mb-4">Deployment Status</h3>
        <DeploymentStatusSummary deployments={data.deployments} />
      </div>
    </div>
  )
}

function DeploymentCard({
  name,
  deployment,
}: {
  name: string
  deployment: any
}) {
  const [isExpanded, setIsExpanded] = useState(false)

  const status = deployment.status
  const replicas = {
    desired: status?.replicas || 0,
    ready: status?.readyReplicas || 0,
    available: status?.availableReplicas || 0,
    updated: status?.updatedReplicas || 0,
  }

  const isHealthy =
    replicas.ready === replicas.desired &&
    replicas.available === replicas.desired

  return (
    <div className="border rounded-lg">
      hey
      <div
        className="p-4 cursor-pointer hover:bg-gray-50"
        onClick={() => setIsExpanded(!isExpanded)}
      >
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <div
              className={`w-3 h-3 rounded-full ${isHealthy ? 'bg-green-400' : 'bg-yellow-400'}`}
            />
            <h4 className="font-medium">{name}</h4>
            <span className="text-sm text-gray-500">
              {replicas.ready}/{replicas.desired} replicas ready
            </span>
          </div>
          <div className="flex items-center space-x-2">
            <StatusBadge status={isHealthy ? 'healthy' : 'pending'} />
            <span className="text-xs text-gray-400">
              {isExpanded ? '▼' : '▶'}
            </span>
          </div>
        </div>
      </div>
      {isExpanded && (
        <div className="border-t bg-gray-50">
          <div className="p-4 space-y-4">
            {/* Replica Status */}
            <div className="grid grid-cols-4 gap-4">
              <div className="text-center">
                <div className="text-lg font-semibold text-blue-600">
                  {replicas.desired}
                </div>
                <div className="text-xs text-gray-500">Desired</div>
              </div>
              <div className="text-center">
                <div className="text-lg font-semibold text-green-600">
                  {replicas.ready}
                </div>
                <div className="text-xs text-gray-500">Ready</div>
              </div>
              <div className="text-center">
                <div className="text-lg font-semibold text-blue-600">
                  {replicas.available}
                </div>
                <div className="text-xs text-gray-500">Available</div>
              </div>
              <div className="text-center">
                <div className="text-lg font-semibold text-purple-600">
                  {replicas.updated}
                </div>
                <div className="text-xs text-gray-500">Updated</div>
              </div>
            </div>

            {/* Conditions */}
            {status?.conditions && (
              <div className="space-y-2">
                <h5 className="font-medium text-sm text-gray-700">
                  Conditions
                </h5>
                {status.conditions.map((condition: any, index: number) => (
                  <div
                    key={index}
                    className="flex items-center justify-between text-sm"
                  >
                    <span className="font-medium">{condition.type}</span>
                    <div className="flex items-center space-x-2">
                      <span
                        className={`px-2 py-1 rounded text-xs ${
                          condition.status === 'True'
                            ? 'bg-green-100 text-green-800'
                            : 'bg-red-100 text-red-800'
                        }`}
                      >
                        {condition.status}
                      </span>
                      <span className="text-gray-500">{condition.reason}</span>
                    </div>
                  </div>
                ))}
              </div>
            )}

            {/* Metadata */}
            <div className="space-y-2">
              <h5 className="font-medium text-sm text-gray-700">Details</h5>
              <KeyValueList
                values={[
                  {
                    key: 'Created',
                    value: new Date(
                      deployment.metadata.creationTimestamp
                    ).toLocaleString(),
                    type: 'string',
                  },
                  {
                    key: 'Generation',
                    value: String(deployment.metadata.generation),
                    type: 'number',
                  },
                  {
                    key: 'Resource Version',
                    value: deployment.metadata.resourceVersion,
                    type: 'string',
                  },
                  {
                    key: 'UID',
                    value: deployment.metadata.uid,
                    type: 'string',
                  },
                ]}
              />
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

function DeploymentsTab({ deployments }: { deployments: Record<string, any> }) {
  return (
    <div className="space-y-4">
      {Object.entries(deployments).map(([namespace, namespaceDeployments]) => (
        <div key={namespace} className="bg-white rounded-lg border">
          <div className="px-6 py-4 border-b bg-gray-50">
            <h3 className="text-lg font-semibold">Namespace: {namespace}</h3>
          </div>

          <div className="p-6 space-y-4">
            {Object.entries(namespaceDeployments).map(
              ([name, deployment]: [string, any]) => (
                <DeploymentCard
                  key={name}
                  name={name}
                  deployment={deployment}
                />
              )
            )}
          </div>
        </div>
      ))}
    </div>
  )
}

function ServicesTab({ services }: { services: Record<string, any> }) {
  return (
    <div className="space-y-4">
      {Object.entries(services).map(([namespace, namespaceServices]) => (
        <div key={namespace} className="bg-white rounded-lg border">
          <div className="px-6 py-4 border-b bg-gray-50">
            <h3 className="text-lg font-semibold">Namespace: {namespace}</h3>
          </div>

          <div className="p-6 space-y-4">
            services
            {Object.entries(namespaceServices).map(
              ([name, service]: [string, any]) => (
                <div key={name} className="border rounded-lg p-4">
                  <div className="flex items-center justify-between mb-3">
                    <h4 className="font-medium">{name}</h4>
                    <StatusBadge status="healthy" />
                  </div>
                  <KeyValueList
                    values={objectToKeyValueArray({
                      Created: new Date(
                        service.metadata.creationTimestamp
                      ).toLocaleString(),
                      UID: service.metadata.uid,
                      'Resource Version': service.metadata.resourceVersion,
                      Type: service.spec?.type || 'Unknown',
                      'Cluster IP': service.spec?.clusterIP || 'None',
                      'Session Affinity':
                        service.spec?.sessionAffinity || 'None',
                    })}
                  />
                </div>
              )
            )}
          </div>
        </div>
      ))}
    </div>
  )
}

function IngressesTab({ ingresses }: { ingresses: Record<string, any> }) {
  const hasIngresses = Object.keys(ingresses).length > 0

  return (
    <div className="bg-white rounded-lg border p-8 text-center">
      <div className="text-gray-400">
        <p>
          {hasIngresses
            ? 'No ingress configuration found'
            : 'No ingresses configured'}
        </p>
      </div>
    </div>
  )
}

function ResourceDetail({ name, resource }: { name: string; resource: any }) {
  return (
    <div className="bg-white rounded-lg border">
      <div className="px-6 py-4 border-b">
        <div className="flex items-center justify-between">
          hey
          <h3 className="text-lg font-semibold">{resource.Kind}</h3>
          <span className="text-sm text-gray-500">{name}</span>
        </div>
      </div>
      <div className="p-6">
        <pre className="bg-gray-900 text-green-400 p-4 rounded-lg overflow-x-auto text-sm font-mono">
          {resource.Content}
        </pre>
      </div>
    </div>
  )
}

function ResourcesTab({ resources }: { resources: Record<string, any> }) {
  const [selectedResource, setSelectedResource] = useState<string | null>(null)

  return (
    <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
      {/* Resource List */}
      <div className="lg:col-span-1">
        <div className="bg-white rounded-lg border">
          <div className="px-4 py-3 border-b bg-gray-50">
            <h3 className="font-medium">
              Resources ({Object.keys(resources).length})
            </h3>
          </div>
          <div className="divide-y">
            {Object.entries(resources).map(
              ([name, resource]: [string, any]) => (
                <button
                  key={name}
                  onClick={() => setSelectedResource(name)}
                  className={`w-full text-left px-4 py-3 hover:bg-gray-50 ${
                    selectedResource === name
                      ? 'bg-blue-50 border-r-2 border-blue-500'
                      : ''
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <div className="min-w-0 flex-1">
                      <div className="font-medium text-sm">{resource.Kind}</div>
                      <div className="text-xs text-gray-500 truncate">
                        {name}
                      </div>
                    </div>
                    <span
                      className={`px-2 py-1 rounded text-xs font-medium ml-2 ${
                        resource.Kind === 'Deployment'
                          ? 'bg-blue-100 text-blue-800'
                          : resource.Kind === 'Service'
                            ? 'bg-green-100 text-green-800'
                            : 'bg-gray-100 text-gray-800'
                      }`}
                    >
                      {resource.Kind}
                    </span>
                  </div>
                </button>
              )
            )}
          </div>
        </div>
      </div>

      {/* Resource Detail */}
      <div className="lg:col-span-2">
        {selectedResource ? (
          <ResourceDetail
            name={selectedResource}
            resource={resources[selectedResource]}
          />
        ) : (
          <div className="bg-white rounded-lg border p-8 text-center">
            <div className="text-gray-400">
              <p>Select a resource to view details</p>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

function ManifestTab({ manifest }: { manifest: string }) {
  const [isCopied, setIsCopied] = useState(false)

  const copyToClipboard = async () => {
    await navigator.clipboard.writeText(manifest)
    setIsCopied(true)
    setTimeout(() => setIsCopied(false), 2000)
  }

  return (
    <div className="bg-white rounded-lg border">
      <div className="px-6 py-4 border-b flex items-center justify-between">
        <h3 className="text-lg font-semibold">Helm Manifest</h3>
        <button
          onClick={copyToClipboard}
          className="flex items-center space-x-2 px-3 py-2 bg-gray-100 hover:bg-gray-200 rounded-lg text-sm transition-colors"
        >
          <span>{isCopied ? 'Copied!' : 'Copy'}</span>
        </button>
      </div>
      <div className="p-6">
        <CodeBlock className="!max-h-fit" language="yml">
          {manifest}
        </CodeBlock>
      </div>
    </div>
  )
}

// Custom Tabs Components
function Tabs({
  value,
  onValueChange,
  children,
}: {
  value: string
  onValueChange: (value: string) => void
  children: React.ReactNode
}) {
  return (
    <div>
      {React.Children.map(children, (child) =>
        React.isValidElement(child)
          ? React.cloneElement(child, { value, onValueChange } as any)
          : child
      )}
    </div>
  )
}

function TabsList({
  children,
  className,
  value,
  onValueChange,
}: {
  children: React.ReactNode
  className?: string
  value?: string
  onValueChange?: (value: string) => void
}) {
  return (
    <div className={`flex space-x-1 rounded-lg bg-gray-100 p-1 ${className}`}>
      {React.Children.map(children, (child) =>
        React.isValidElement(child)
          ? React.cloneElement(child, { value, onValueChange } as any)
          : child
      )}
    </div>
  )
}

function TabsTrigger({
  triggerValue,
  children,
  className,
  value,
  onValueChange,
}: {
  triggerValue: string
  children: React.ReactNode
  className?: string
  value?: string
  onValueChange?: (value: string) => void
}) {
  const isActive = value === triggerValue

  return (
    <button
      onClick={() => onValueChange?.(triggerValue)}
      className={`px-3 py-2 text-sm font-medium rounded-md transition-colors ${
        isActive
          ? 'bg-white text-gray-900 shadow-sm'
          : 'text-gray-600 hover:text-gray-900 hover:bg-gray-50'
      } ${className}`}
    >
      {children}
    </button>
  )
}

function TabsContent({
  contentValue,
  children,
  value,
}: {
  contentValue: string
  children: React.ReactNode
  value?: string
}) {
  if (value !== contentValue) return null

  return <div className="mt-6">{children}</div>
}

// Main Component
export function HelmDeploymentViewer({ data }: { data: HelmDeploymentData }) {
  const [activeTab, setActiveTab] = useState('overview')

  const deploymentCount = Object.keys(data.deployments).length
  const serviceCount = Object.keys(data.services).length
  const ingressCount = Object.keys(data.ingresses).length
  const resourceCount = Object.keys(data.resources).length

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-7xl mx-auto space-y-6">
        {/* Header with Release Info */}
        <DeploymentHeader data={data} />

        <Tabs value={activeTab} onValueChange={setActiveTab}>
          <TabsList className="grid w-full grid-cols-6">
            <TabsTrigger triggerValue="overview">Overview</TabsTrigger>
            <TabsTrigger triggerValue="deployments" className="relative">
              Deployments
              {deploymentCount > 0 && (
                <span className="ml-1 bg-blue-100 text-blue-800 text-xs px-1.5 py-0.5 rounded-full">
                  {deploymentCount}
                </span>
              )}
            </TabsTrigger>
            <TabsTrigger triggerValue="services" className="relative">
              Services
              {serviceCount > 0 && (
                <span className="ml-1 bg-green-100 text-green-800 text-xs px-1.5 py-0.5 rounded-full">
                  {serviceCount}
                </span>
              )}
            </TabsTrigger>
            <TabsTrigger triggerValue="ingresses">Ingresses</TabsTrigger>
            <TabsTrigger triggerValue="resources">Resources</TabsTrigger>
            <TabsTrigger triggerValue="manifest">Manifest</TabsTrigger>
          </TabsList>

          <TabsContent contentValue="overview">
            <OverviewTab data={data} />
          </TabsContent>

          <TabsContent contentValue="deployments">
            <DeploymentsTab deployments={data.deployments} />
          </TabsContent>

          <TabsContent contentValue="services">
            <ServicesTab services={data.services} />
          </TabsContent>

          <TabsContent contentValue="ingresses">
            <IngressesTab ingresses={data.ingresses} />
          </TabsContent>

          <TabsContent contentValue="resources">
            <ResourcesTab resources={data.resources} />
          </TabsContent>

          <TabsContent contentValue="manifest">
            <ManifestTab manifest={data.manifest} />
          </TabsContent>
        </Tabs>
      </div>
    </div>
  )
}
