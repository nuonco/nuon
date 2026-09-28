import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getInstallStatus } from '@/lib'
import { INSTALL_STATUS_AXES, InstallStatusAxis } from './InstallStatusSummary'

export const InstallStatusAxisCell = ({ installId }: { installId: string }) => {
  const { org } = useOrg()

  const { data, isLoading, isError } = useQuery({
    queryKey: ['install-status', org?.id, installId],
    queryFn: () => getInstallStatus({ orgId: org.id, installId }),
    enabled: !!org?.id && !!installId,
    placeholderData: keepPreviousData,
    refetchInterval: 20000,
  })

  return (
    <div className="flex w-max flex-col items-start gap-1.5">
      {INSTALL_STATUS_AXES.map(({ key, label }) => (
        <span key={key} title={label}>
          <InstallStatusAxis
            axis={data?.[key]}
            loading={isLoading && !data}
            unavailable={isError && !data}
          />
        </span>
      ))}
    </div>
  )
}
