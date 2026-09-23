import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { ReprovisionStackButton } from '@/components/installs/management/ReprovisionStack'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallStack } from '@/lib'
import { InstallStackVersions } from './InstallStackVersions'

const byCreatedAtDesc = (
  a: { created_at?: string },
  b: { created_at?: string }
) => {
  const aTime = a.created_at ? Date.parse(a.created_at) : 0
  const bTime = b.created_at ? Date.parse(b.created_at) : 0
  return bTime - aTime
}

export const InstallStackVersionsContainer = ({
  pollInterval = 20000,
  shouldPoll = true,
}: {
  pollInterval?: number
  shouldPoll?: boolean
}) => {
  const { org } = useOrg()
  const { install } = useInstall()

  const { data: stack, isLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['install-stack', org?.id, install?.id],
    queryFn: () => getInstallStack({ orgId: org.id, installId: install.id }),
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!install?.id,
  })

  const versions = [...(stack?.versions ?? [])].sort(byCreatedAtDesc)

  return (
    <InstallStackVersions
      versions={versions}
      loading={isLoading}
      latestAction={
        versions.length ? (
          <ReprovisionStackButton size="sm" variant="secondary" />
        ) : null
      }
    />
  )
}
