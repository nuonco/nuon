import { useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { SyncedFilterContainer } from '@/components/common/SyncedFilter'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useSyncedOnlyFilter } from '@/hooks/use-synced-only-filter'
import { getInstallRunbooks } from '@/lib'
import {
  InstallRunbooksTable,
  parseInstallRunbooksToTableData,
} from './InstallRunbooksTable'

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
  const { syncedOnly } = useSyncedOnlyFilter()

  const { data: result, isLoading } = useQuery({
    queryKey: ['install-runbooks', org?.id, install?.id, offset, q],
    queryFn: () =>
      getInstallRunbooks({
        orgId: org!.id,
        installId: install!.id,
        offset,
        limit: LIMIT,
        q,
      }),
    placeholderData: keepPreviousData,
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!install?.id,
  })

  const { data: removedResult } = useQuery({
    queryKey: ['install-runbooks-removed', org?.id, install?.id, q],
    queryFn: () =>
      getInstallRunbooks({
        orgId: org!.id,
        installId: install!.id,
        offset: 0,
        limit: 100,
        q,
        synced: false,
      }),
    placeholderData: keepPreviousData,
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!install?.id && !syncedOnly,
  })

  const removedRunbooks =
    offset === 0 && !syncedOnly ? (removedResult?.data ?? []) : []

  const removedRows = parseInstallRunbooksToTableData(
    removedRunbooks,
    org?.id ?? '',
    install?.id ?? '',
    labelColors,
    true
  )
  const currentRows = parseInstallRunbooksToTableData(
    result?.data ?? [],
    org?.id ?? '',
    install?.id ?? '',
    labelColors
  )

  return (
    <InstallRunbooksTable
      data={[...removedRows, ...currentRows]}
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
