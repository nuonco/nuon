import { useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { LabelFilterDropdown } from '@/components/common/LabelFilterDropdown'
import { useOrg } from '@/hooks/use-org'
import { getInstalls, getInstallLabelKeys } from '@/lib'
import { CreateInstallButton } from '../CreateInstall'
import { InstallsTable, parseInstallsToTableData } from './InstallsTable'

const LIMIT = 20

export const InstallsTableContainer = ({
  pollInterval = 20000,
  shouldPoll = true,
}: {
  pollInterval?: number
  shouldPoll?: boolean
} = {}) => {
  const [searchParams] = useSearchParams()
  const { org } = useOrg()
  const offset = Number(searchParams.get('offset') ?? 0)

  const { data: result, isLoading } = useQuery({
    queryKey: ['installs', org.id, offset, searchParams.get('q'), searchParams.get('labels')],
    queryFn: () =>
      getInstalls({
        orgId: org.id,
        offset,
        limit: LIMIT,
        q: searchParams.get('q') || undefined,
        labels: searchParams.get('labels') || undefined,
      }),
    placeholderData: keepPreviousData,
    refetchInterval: shouldPoll ? pollInterval : false,
  })

  return (
    <InstallsTable
      data={parseInstallsToTableData(result?.data ?? [], org.id)}
      isLoading={isLoading}
      emptyStateAction={<CreateInstallButton />}
      filterActions={
        <div className="flex items-center gap-3">
          <LabelFilterDropdown
            queryKey={['install-label-keys', org.id]}
            queryFn={() => getInstallLabelKeys({ orgId: org.id })}
          />
          <CreateInstallButton
            className="!w-full !flex !justify-center md:!w-fit"
            variant="primary"
          />
        </div>
      }
      pagination={{
        hasNext: result?.pagination?.hasNext ?? false,
        offset,
        limit: LIMIT,
      }}
    />
  )
}
