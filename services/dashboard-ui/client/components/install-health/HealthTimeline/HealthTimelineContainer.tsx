import { useQuery } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import {
  getInstallComponentHealthTimeline,
  getInstallHealthTimeline,
} from '@/lib'
import type {
  TInstallComponentHealthTimeline,
  TInstallHealthTimeline,
} from '@/types'
import { HealthCardActions } from '@/components/install-health/HealthCardActions'
import { HealthTimeline } from './HealthTimeline'

export const HealthTimelineContainer = ({
  installComponentId,
  days = 90,
  pollInterval = 20000,
  shouldPoll = false,
}: {
  installComponentId?: string
  days?: number
  pollInterval?: number
  shouldPoll?: boolean
}) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const isComponentScope = !!installComponentId

  const { data: installTimeline, isLoading: isInstallLoading } = useQuery({
    queryKey: ['install-health-timeline', org?.id, install?.id, days],
    queryFn: () =>
      getInstallHealthTimeline({ orgId: org!.id, installId: install!.id, days }),
    enabled: !!org?.id && !!install?.id && !isComponentScope,
    refetchInterval: shouldPoll ? pollInterval : false,
  })

  const { data: componentTimeline, isLoading: isComponentLoading } = useQuery({
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
      currentHealth={timeline?.current_health}
      components={
        isComponentScope
          ? undefined
          : (timeline as TInstallHealthTimeline | undefined)?.components
      }
      componentBasePath={`/${org?.id}/installs/${install?.id}/components`}
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
