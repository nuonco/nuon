import { useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { useOrg } from '@/hooks/use-org'
import { getApps } from '@/lib'
import { AppsTable, parseAppsToTableData } from './AppsTable'

const LIMIT = 20

export const AppsTableContainer = ({
  pollInterval = 15000,
  shouldPoll = true,
}: {
  pollInterval?: number
  shouldPoll?: boolean
} = {}) => {
  const [searchParams] = useSearchParams()
  const { org } = useOrg()
  const offset = Number(searchParams.get('offset') ?? 0)

  const { data: result, isLoading } = useQuery({
    queryKey: ['apps', org.id, offset, searchParams.get('q')],
    queryFn: () => getApps({
      orgId: org.id,
      offset,
      limit: LIMIT,
      q: searchParams.get('q') || undefined,
    }),
    placeholderData: keepPreviousData,
    refetchInterval: shouldPoll ? pollInterval : false,
  })

  return (
    <AppsTable
      data={parseAppsToTableData(result?.data ?? [], org.id)}
      isLoading={isLoading}
      emptyStateAction={
        <Button href="/onboarding">
          <Icon size="14" variant="AppWindow" />
          Create app
        </Button>
      }
      pagination={{ hasNext: result?.pagination?.hasNext ?? false, offset, limit: LIMIT }}
    />
  )
}
