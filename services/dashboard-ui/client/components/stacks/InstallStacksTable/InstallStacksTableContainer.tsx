import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallStack } from '@/lib'
import { InstallStacksTable } from './InstallStacksTable'

export const InstallStacksTableContainer = ({
  pollInterval = 20000,
  shouldPoll,
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

  return (
    <InstallStacksTable
      versions={stack?.versions ?? []}
      orgId={org?.id}
      appId={install?.app_id}
      isLoading={isLoading}
    />
  )
}
