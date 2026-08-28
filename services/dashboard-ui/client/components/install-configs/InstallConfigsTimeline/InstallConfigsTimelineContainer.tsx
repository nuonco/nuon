import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallConfigVersions } from '@/lib'
import { InstallConfigsTimeline } from './InstallConfigsTimeline'

export const InstallConfigsTimelineContainer = ({
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
    queryKey: ['install-config-versions', org?.id, install?.id],
    queryFn: () =>
      getInstallConfigVersions({ orgId: org!.id, installId: install!.id }),
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!install?.id,
  })

  return (
    <InstallConfigsTimeline
      versions={versions ?? []}
      isLoading={isLoading}
      orgId={org?.id}
      installId={install?.id}
    />
  )
}
