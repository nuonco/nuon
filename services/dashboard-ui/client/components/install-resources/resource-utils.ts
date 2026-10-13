import type { TInstallResource } from '@/types'
import { resourceStatusFields } from './InstallResourceDetailPanel/resource-details'

export function isHealthCheckResource(resource: TInstallResource): boolean {
  return resource.provider === 'probe' || resource.provider === 'custom'
}

export function resourceSummary(
  resource: TInstallResource
): string | undefined {
  const fields = resourceStatusFields(resource)
  return fields.length
    ? fields
        .slice(0, 3)
        .map(([label, value]) =>
          label === 'Ready' &&
          [
            'Pod',
            'Deployment',
            'ReplicaSet',
            'StatefulSet',
            'DaemonSet',
          ].includes(resource.kind || '')
            ? `${value} ready`
            : label === 'Restarts'
              ? `${value} restarts`
              : [
                    'Status',
                    'Phase',
                    'phase',
                    'Type',
                    'Result',
                    'Latency',
                  ].includes(label)
                ? value
                : `${label}=${value}`
        )
        .join(' · ')
    : undefined
}

export function resourceIdentity(resource: TInstallResource): string {
  return JSON.stringify([
    resource.install_id || '',
    resource.source || 'component',
    resource.install_component_id || resource.component_id || '',
    resource.owner_name || '',
    resource.provider || '',
    resource.api_group || '',
    resource.kind || '',
    resource.namespace || '',
    resource.name || '',
  ])
}

export function resourceOwner(
  resource: TInstallResource,
  componentNames: Record<string, string>
): { key: string; label: string } {
  if (resource.source === 'sandbox') {
    return {
      key: `sandbox:${resource.owner_name || ''}`,
      label: `Sandbox · ${resource.owner_name || 'Unknown release'}`,
    }
  }
  const id = resource.install_component_id || resource.component_id || ''
  return {
    key: `component:${id}`,
    label: componentNames[id] || id || 'Unknown component',
  }
}
