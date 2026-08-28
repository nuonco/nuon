import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getAppInstallSyncs } from '@/lib'
import { AppInstallSyncsTimeline } from './AppInstallSyncsTimeline'

export const AppInstallSyncsTimelineContainer = ({
  pollInterval = 10000,
  shouldPoll,
}: {
  pollInterval?: number
  shouldPoll?: boolean
}) => {
  const { org } = useOrg()
  const { app } = useApp()

  const { data: syncs, isLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-install-syncs', org?.id, app?.id],
    queryFn: () => getAppInstallSyncs({ appId: app!.id, orgId: org!.id }),
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!app?.id,
  })

  return (
    <AppInstallSyncsTimeline
      syncs={syncs ?? []}
      isLoading={isLoading}
      orgId={org?.id}
      appId={app?.id}
    />
  )
}
