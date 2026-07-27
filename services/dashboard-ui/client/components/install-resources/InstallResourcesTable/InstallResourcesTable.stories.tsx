export default {
  title: 'InstallResources/InstallResourcesTable',
}

import { useState } from 'react'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import type { TInstallResource } from '@/types'
import {
  groupComponentResources,
  groupSandboxResources,
  InstallResourcesTable,
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

const componentNames = { 'instcmp-1': 'web', 'instcmp-2': 'database' }

const InstallResourcesTableStory = ({
  resources,
}: {
  resources: TInstallResource[]
}) => {
  const [kind, setKind] = useState('')
  const [namespace, setNamespace] = useState('')
  const [health, setHealth] = useState('')

  const filtered = resources.filter((r) => {
    if (kind && r.kind !== kind) return false
    if (namespace && r.namespace !== namespace) return false
    if (health && r.health !== health) return false
    return true
  })

  return (
    <SurfacesProvider>
      <InstallResourcesTable
        componentGroups={groupComponentResources(filtered, componentNames)}
        sandboxGroups={groupSandboxResources(filtered)}
        isLoading={false}
        kind={kind}
        namespace={namespace}
        health={health}
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

export const Default = () => (
  <InstallResourcesTableStory resources={mockResources} />
)

export const Loading = () => (
  <InstallResourcesTable
    componentGroups={[]}
    sandboxGroups={[]}
    isLoading
    kind=""
    namespace=""
    health=""
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
    isLoading={false}
    kind=""
    namespace=""
    health=""
    kindOptions={[]}
    namespaceOptions={[]}
    onKindChange={() => {}}
    onNamespaceChange={() => {}}
    onHealthChange={() => {}}
  />
)
