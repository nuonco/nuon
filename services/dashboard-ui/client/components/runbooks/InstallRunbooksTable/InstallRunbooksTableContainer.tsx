import { useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { SyncedFilterContainer } from '@/components/common/SyncedFilter'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallRunbooks } from '@/lib'
import { InstallRunbooksTable, parseInstallRunbooksToTableData } from './InstallRunbooksTable'

const LIMIT = 20

export const InstallRunbooksTableContainer = ({
  pollInterval = 20000,
  shouldPoll,
}: {
  pollInterval?: number
  shouldPoll?: boolean
} = {}) => {
  const [searchParams] = useSearchParams()
  const { org } = useOrg()
  const { install, labelColors } = useInstall()
  const offset = Number(searchParams.get('offset') ?? 0)
  const q = searchParams.get('q') || undefined
  const synced = searchParams.get('synced') === 'false' ? false : undefined

  const { data: result, isLoading } = useQuery({
    queryKey: ['install-runbooks', org?.id, install?.id, offset, q, synced],
    queryFn: () =>
      getInstallRunbooks({
        orgId: org!.id,
        installId: install!.id,
        offset,
        limit: LIMIT,
        q,
        synced,
      }),
    placeholderData: keepPreviousData,
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!install?.id,
  })

  return (
    <InstallRunbooksTable
      data={parseInstallRunbooksToTableData(
        result?.data ?? [],
        org?.id ?? '',
        install?.id ?? '',
        labelColors
      )}
      isLoading={isLoading}
      filterActions={<SyncedFilterContainer />}
      pagination={{
        hasNext: result?.pagination?.hasNext ?? false,
        offset,
        limit: LIMIT,
      }}
    />
  )
}
