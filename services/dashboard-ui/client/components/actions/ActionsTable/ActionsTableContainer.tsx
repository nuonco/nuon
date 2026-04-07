import { useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getActions } from '@/lib'
import { ActionsTable, parseActionsToTableData } from './ActionsTable'

const LIMIT = 20

export const ActionsTableContainer = ({
  pollInterval = 20000,
  shouldPoll = true,
}: {
  pollInterval?: number
  shouldPoll?: boolean
} = {}) => {
  const [searchParams] = useSearchParams()
  const { org } = useOrg()
  const { app } = useApp()
  const offset = Number(searchParams.get('offset') ?? 0)

  const { data: result, isLoading } = useQuery({
    queryKey: ['actions', org?.id, app?.id, offset, searchParams.get('q')],
    queryFn: () =>
      getActions({
        orgId: org.id,
        appId: app.id,
        offset,
        limit: LIMIT,
        q: searchParams.get('q') || undefined,
      }),
    placeholderData: keepPreviousData,
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!app?.id,
  })

  return (
    <ActionsTable
      data={parseActionsToTableData(result?.data ?? [], org.id, app.id)}
      isLoading={isLoading}
      pagination={{
        hasNext: result?.pagination?.hasNext ?? false,
        offset,
        limit: LIMIT,
      }}
    />
  )
}
