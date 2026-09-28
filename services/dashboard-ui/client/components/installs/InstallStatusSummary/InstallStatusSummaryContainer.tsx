import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallStatus } from '@/lib'
import { InstallStatusSummary } from './InstallStatusSummary'

export const InstallStatusSummaryContainer = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  const { data, isLoading, isError } = useQuery({
    queryKey: ['install-status', org?.id, install?.id],
    queryFn: () => getInstallStatus({ orgId: org.id, installId: install.id }),
    enabled: !!org?.id && !!install?.id,
    placeholderData: keepPreviousData,
    refetchInterval: 20000,
  })

  return (
    <InstallStatusSummary
      status={data}
      loading={isLoading && !data}
      unavailable={isError && !data}
    />
  )
}
