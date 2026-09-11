import type { TCloudPlatform, TInstall } from '@/types'
import { humanize } from '@/utils/string-utils'

export interface IInstallStatusFacet {
  id: string
  title: string
  icon:
    | 'SneakerMoveIcon'
    | 'ShippingContainerIcon'
    | 'CardsIcon'
    | 'HeartbeatIcon'
    | 'FileDashedIcon'
  status: string
  description?: string
}

const STALE_PHASE_STATUSES = new Set([
  'active',
  'pending',
  'executing',
  'queued',
  'planning',
  'syncing',
])

const phaseAdjusted = (status?: string, phase?: string) => {
  if (!status || !STALE_PHASE_STATUSES.has(status)) return status
  if (phase === 'deprovisioned' || phase === 'deprovisioning') return phase
  return status
}

const facetTitle = (label: string, status: string) =>
  `${label} ${humanize(status).toLowerCase()}`

export const installStatusFacets = (
  install?: Partial<TInstall>
): IInstallStatusFacet[] => {
  const phase = install?.lifecycle_phase?.phase
  const adjust = (status?: string) => phaseAdjusted(status, phase)
  const runnerStatus = adjust(install?.runner_status) ?? 'unknown'
  const sandboxBase = adjust(install?.sandbox_status)
  const sandboxStatus =
    install?.sandbox_health_status && sandboxBase === 'active'
      ? install.sandbox_health_status
      : (sandboxBase ?? 'unknown')
  const componentStatus =
    adjust(install?.composite_component_status) ?? 'unknown'

  const facets: IInstallStatusFacet[] = [
    {
      id: 'runner',
      title: facetTitle('Runner', runnerStatus),
      icon: 'SneakerMoveIcon',
      status: runnerStatus,
      description: install?.runner_status_description,
    },
    {
      id: 'sandbox',
      title: facetTitle('Sandbox', sandboxStatus),
      icon: 'ShippingContainerIcon',
      status: sandboxStatus,
      description:
        sandboxStatus === sandboxBase
          ? install?.sandbox_status_description
          : (install?.sandbox_health_message ??
            install?.sandbox_status_description),
    },
    {
      id: 'components',
      title: facetTitle('Components', componentStatus),
      icon: 'CardsIcon',
      status: componentStatus,
      description: install?.composite_component_status_description,
    },
  ]

  if (install?.composite_health_status) {
    facets.push({
      id: 'health',
      title: facetTitle('Health', install.composite_health_status),
      icon: 'HeartbeatIcon',
      status: install.composite_health_status,
      description: install?.composite_health_status_description,
    })
  }

  if (install?.drifted_objects) {
    const drifted = install.drifted_objects.length
    facets.push({
      id: 'drift',
      title: drifted ? 'Drift detected' : 'No drift',
      icon: 'FileDashedIcon',
      status: drifted ? 'warn' : 'active',
      description: drifted
        ? `${drifted} resource${drifted === 1 ? '' : 's'} have drifted from the last applied state.`
        : undefined,
    })
  }

  return facets
}

export const installCloudLocation = (install?: Partial<TInstall>) => {
  const platform = (install?.cloud_platform ?? '').toLowerCase()
  const cloudPlatform: TCloudPlatform =
    platform === 'aws' || platform === 'azure' || platform === 'gcp'
      ? platform
      : 'unknown'

  return {
    platform: cloudPlatform,
    region:
      cloudPlatform === 'azure'
        ? undefined
        : (install?.aws_account?.region ?? install?.gcp_account?.region),
    location:
      cloudPlatform === 'azure' ? install?.azure_account?.location : undefined,
  }
}
