import { useMemo } from 'react'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import {
  getInstallComponentHealthTimeline,
  getInstallHealthTimeline,
} from '@/lib'
import type {
  TComponentType,
  TInstallComponent,
  TInstallComponentHealthTimeline,
  TInstallHealthTimeline,
} from '@/types'
import { HealthCardActions } from '@/components/install-health/HealthCardActions'
import {
  HealthTimeline,
  type THealthTimelineComponentLink,
} from './HealthTimeline'

export const HealthTimelineContainer = ({
  installComponentId,
  days = 90,
  pollInterval = 20000,
  shouldPoll = false,
  componentBasePath,
  getComponentHref,
  groupByKind = false,
}: {
  installComponentId?: string
  days?: number
  pollInterval?: number
  shouldPoll?: boolean
  componentBasePath?: string
  getComponentHref?: (component: THealthTimelineComponentLink) => string
  groupByKind?: boolean
}) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const isComponentScope = !!installComponentId

  const { data: installTimeline, isLoading: isInstallLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['install-health-timeline', org?.id, install?.id, days],
    queryFn: () =>
      getInstallHealthTimeline({ orgId: org!.id, installId: install!.id, days }),
    enabled: !!org?.id && !!install?.id && !isComponentScope,
    refetchInterval: shouldPoll ? pollInterval : false,
  })

  const { data: componentTimeline, isLoading: isComponentLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: [
      'install-component-health-timeline',
      org?.id,
      install?.id,
      installComponentId,
      days,
    ],
    queryFn: () =>
      getInstallComponentHealthTimeline({
        orgId: org!.id,
        installId: install!.id,
        componentId: installComponentId!,
        days,
      }),
    enabled: !!org?.id && !!install?.id && isComponentScope,
    refetchInterval: shouldPoll ? pollInterval : false,
  })

  const timeline = isComponentScope ? componentTimeline : installTimeline
  const isLoading = isComponentScope ? isComponentLoading : isInstallLoading

  const componentTypes = useMemo(() => {
    const types = new Map<string, TComponentType>()
    for (const installComponent of (install?.install_components ??
      []) as TInstallComponent[]) {
      const type = installComponent.component?.type
      if (!type) continue
      if (installComponent.component_id)
        types.set(installComponent.component_id, type)
      if (installComponent.id) types.set(installComponent.id, type)
    }
    return types
  }, [install?.install_components])

  const components = isComponentScope
    ? undefined
    : (timeline as TInstallHealthTimeline | undefined)?.components?.map(
        (component) => ({
          ...component,
          component_type:
            componentTypes.get(component.component_id ?? '') ??
            componentTypes.get(component.install_component_id),
        })
      )

  return (
    <HealthTimeline
      headerAction={
        !isComponentScope && install?.id ? (
          <HealthCardActions installId={install.id} />
        ) : undefined
      }
      scope={isComponentScope ? 'component' : 'install'}
      days={timeline?.days ?? days}
      daily={timeline?.daily}
      uptimePercent={timeline?.uptime_percent}
      observedSeconds={timeline?.observed_seconds}
      clusterAccessError={
        isComponentScope
          ? undefined
          : (timeline as TInstallHealthTimeline | undefined)?.cluster_access_error
      }
      currentHealth={timeline?.current_health}
      components={components}
      groupByKind={groupByKind}
      componentBasePath={
        componentBasePath ??
        `/${org?.id}/installs/${install?.id}/components`
      }
      getComponentHref={getComponentHref}
      transitions={
        isComponentScope
          ? (timeline as TInstallComponentHealthTimeline | undefined)
              ?.transitions
          : undefined
      }
      deployBasePath={
        isComponentScope
          ? `/${org?.id}/installs/${install?.id}/components/${installComponentId}/deploys`
          : undefined
      }
      isLoading={isLoading}
    />
  )
}
