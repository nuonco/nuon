import { useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getInstalls } from '@/lib'
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
    queryKey: ['installs', org.id, offset, searchParams.get('q')],
    queryFn: () =>
      getInstalls({
        orgId: org.id,
        offset,
        limit: LIMIT,
        q: searchParams.get('q') || undefined,
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
        <CreateInstallButton
          className="!w-full !flex !justify-center md:!w-fit"
          variant="primary"
        />
      }
      pagination={{
        hasNext: result?.pagination?.hasNext ?? false,
        offset,
        limit: LIMIT,
      }}
    />
  )
}
