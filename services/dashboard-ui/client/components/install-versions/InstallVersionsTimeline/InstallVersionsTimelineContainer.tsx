import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallAppConfigVersions } from '@/lib'
import { InstallVersionsTimeline } from './InstallVersionsTimeline'

export const InstallVersionsTimelineContainer = ({
  pollInterval = 20000,
  shouldPoll,
}: {
  pollInterval?: number
  shouldPoll?: boolean
}) => {
  const { org } = useOrg()
  const { install } = useInstall()

  const { data: versions, isLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['install-app-config-versions', org?.id, install?.id],
    queryFn: () =>
      getInstallAppConfigVersions({ orgId: org!.id, installId: install!.id }),
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!install?.id,
  })

  return (
    <InstallVersionsTimeline
      versions={versions ?? []}
      isLoading={isLoading}
      orgId={org?.id}
      installId={install?.id}
      appId={install?.app_id}
    />
  )
}
