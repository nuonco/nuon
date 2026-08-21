export default {
  title: 'Install Resources/InstallResourcesTable',
}

import { useState } from 'react'
import { useSearchParams } from 'react-router'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import type { TInstallResource } from '@/types'
import {
  groupComponentResources,
  groupSandboxResources,
  healthFacetCounts,
  InstallResourcesTable,
  matchesHealthFilter,
  matchesResourceSearch,
} from './InstallResourcesTable'

const mockResources: TInstallResource[] = [
  {
    install_component_id: 'instcmp-1',
    component_id: 'component-1',
    source: 'component',
    kind: 'Deployment',
    namespace: 'default',
    name: 'web-app',
    health: 'healthy',
    message: '',
    provider: 'kubernetes',
    observed_at: new Date().toISOString(),
  },
  {
    install_component_id: 'instcmp-1',
    component_id: 'component-1',
    source: 'component',
    kind: 'Service',
    namespace: 'default',
    name: 'web-app-svc',
    health: 'progressing',
    message: 'Waiting for endpoints to become ready.',
    provider: 'kubernetes',
    observed_at: new Date().toISOString(),
  },
  {
    install_component_id: 'instcmp-2',
    component_id: 'component-2',
    kind: 'StatefulSet',
    namespace: 'data',
    name: 'postgres',
    health: 'degraded',
    message: 'Pod postgres-0 is CrashLoopBackOff.',
    provider: 'kubernetes',
    observed_at: new Date().toISOString(),
  },
  {
    source: 'sandbox',
    owner_name: 'external-dns',
    kind: 'Deployment',
    namespace: 'external-dns',
    name: 'external-dns',
    health: 'healthy',
    message: '',
    provider: 'kubernetes',
    observed_at: new Date().toISOString(),
  },
  {
    source: 'sandbox',
    owner_name: 'cert-manager',
    kind: 'Deployment',
    namespace: 'cert-manager',
    name: 'cert-manager',
    health: 'healthy',
    message: '',
    provider: 'kubernetes',
    observed_at: new Date().toISOString(),
  },
  {
    source: 'sandbox',
    owner_name: 'cert-manager',
    kind: 'Deployment',
    namespace: 'cert-manager',
    name: 'cert-manager-webhook',
    health: 'degraded',
    message: 'Webhook deployment has 0/1 ready replicas.',
    provider: 'kubernetes',
    observed_at: new Date().toISOString(),
  },
  {
    source: 'sandbox',
    owner_name: 'ingress-nginx',
    kind: 'DaemonSet',
    namespace: 'ingress-nginx',
    name: 'ingress-nginx-controller',
    health: 'healthy',
    message: '',
    provider: 'kubernetes',
    observed_at: new Date().toISOString(),
  },
]

const componentNames = {
  'instcmp-1': 'web',
  'instcmp-2': 'database',
  'instcmp-3': 'config',
  'instcmp-4': 'iam',
  'instcmp-5': 'workers',
  'instcmp-6': 'payments',
}

const InstallResourcesTableStory = ({
  resources,
}: {
  resources: TInstallResource[]
}) => {
  const [kind, setKind] = useState('')
  const [namespace, setNamespace] = useState('')
  const [health, setHealth] = useState('')
  const [searchParams] = useSearchParams()
  const search = searchParams.get('q') ?? ''

  const scoped = resources.filter((r) => {
    if (kind && r.kind !== kind) return false
    if (namespace && r.namespace !== namespace) return false
    return matchesResourceSearch(r, search)
  })
  const filtered = scoped.filter((r) => matchesHealthFilter(r, health))

  return (
    <SurfacesProvider>
      <InstallResourcesTable
        componentGroups={groupComponentResources(filtered, componentNames)}
        sandboxGroups={groupSandboxResources(filtered)}
        healthCounts={healthFacetCounts(scoped)}
        isLoading={false}
        kind={kind}
        namespace={namespace}
        health={health}
        search={search}
        kindOptions={Array.from(
          new Set(resources.map((r) => r.kind).filter((v): v is string => !!v))
        ).sort()}
        namespaceOptions={Array.from(
          new Set(
            resources.map((r) => r.namespace).filter((v): v is string => !!v)
          )
        ).sort()}
        onKindChange={setKind}
        onNamespaceChange={setNamespace}
        onHealthChange={setHealth}
      />
    </SurfacesProvider>
  )
}

const noSignalResources: TInstallResource[] = [
  {
    install_component_id: 'instcmp-3',
    component_id: 'component-3',
    source: 'component',
    kind: 'ConfigMap',
    namespace: 'default',
    name: 'app-settings',
    health: 'not-applicable',
    message: '',
    provider: 'kubernetes',
    observed_at: new Date().toISOString(),
  },
  {
    install_component_id: 'instcmp-3',
    component_id: 'component-3',
    source: 'component',
    kind: 'Job',
    namespace: 'default',
    name: 'migrate',
    health: 'unknown',
    message: 'No probe could be run against this resource.',
    provider: 'kubernetes',
    observed_at: new Date().toISOString(),
  },
  {
    install_component_id: 'instcmp-4',
    component_id: 'component-4',
    source: 'component',
    kind: 'aws_iam_role',
    namespace: '',
    name: 'acme-app-role',
    health: 'unknown',
    message: '',
    provider: 'aws',
    observed_at: new Date().toISOString(),
  },
  {
    install_component_id: 'instcmp-1',
    component_id: 'component-1',
    source: 'component',
    kind: 'Secret',
    namespace: 'default',
    name: 'web-app-tls',
    health: 'unknown',
    message: '',
    provider: 'kubernetes',
    observed_at: new Date().toISOString(),
    removed_from_config: true,
  },
]

const largeHealthyGroup: TInstallResource[] = Array.from(
  { length: 12 },
  (_, idx) => ({
    install_component_id: 'instcmp-5',
    component_id: 'component-5',
    source: 'component',
    kind: 'Pod',
    namespace: 'workers',
    name: `worker-${idx}`,
    health: 'healthy',
    message: '',
    provider: 'kubernetes',
    observed_at: new Date().toISOString(),
  })
)

const staleGroup: TInstallResource[] = [
  {
    install_component_id: 'instcmp-6',
    component_id: 'component-6',
    source: 'component',
    kind: 'Deployment',
    namespace: 'payments',
    name: 'payments-api',
    health: 'healthy',
    message: '',
    provider: 'kubernetes',
    observed_at: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString(),
  },
  {
    install_component_id: 'instcmp-6',
    component_id: 'component-6',
    source: 'component',
    kind: 'Service',
    namespace: 'payments',
    name: 'payments-api-svc',
    health: 'healthy',
    message: '',
    provider: 'kubernetes',
    observed_at: new Date(Date.now() - 3 * 60 * 60 * 1000).toISOString(),
  },
]

export const Default = () => (
  <InstallResourcesTableStory resources={mockResources} />
)

export const WithStaleGroup = () => (
  <InstallResourcesTableStory resources={[...mockResources, ...staleGroup]} />
)

export const WithNoSignalResources = () => (
  <InstallResourcesTableStory
    resources={[...mockResources, ...noSignalResources]}
  />
)

export const WithLargeHealthyGroup = () => (
  <InstallResourcesTableStory
    resources={[...mockResources, ...noSignalResources, ...largeHealthyGroup]}
  />
)

export const Loading = () => (
  <InstallResourcesTable
    componentGroups={[]}
    sandboxGroups={[]}
    healthCounts={{}}
    isLoading
    kind=""
    namespace=""
    health=""
    search=""
    kindOptions={[]}
    namespaceOptions={[]}
    onKindChange={() => {}}
    onNamespaceChange={() => {}}
    onHealthChange={() => {}}
  />
)

export const Empty = () => (
  <InstallResourcesTable
    componentGroups={[]}
    sandboxGroups={[]}
    healthCounts={{}}
    isLoading={false}
    kind=""
    namespace=""
    health=""
    search=""
    kindOptions={[]}
    namespaceOptions={[]}
    onKindChange={() => {}}
    onNamespaceChange={() => {}}
    onHealthChange={() => {}}
  />
)
