import { useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getRunbooks } from '@/lib'
import { RunbooksTable, parseRunbooksToTableData } from './RunbooksTable'

const LIMIT = 20

export const RunbooksTableContainer = () => {
  const [searchParams] = useSearchParams()
  const { org } = useOrg()
  const { app } = useApp()
  const offset = Number(searchParams.get('offset') ?? 0)

  const { data: result, isLoading } = useQuery({
    queryKey: ['runbooks', org?.id, app?.id, offset],
    queryFn: () =>
      getRunbooks({
        orgId: org!.id,
        appId: app!.id,
        offset,
        limit: LIMIT,
      }),
    placeholderData: keepPreviousData,
    enabled: !!org?.id && !!app?.id,
  })

  return (
    <RunbooksTable
      data={parseRunbooksToTableData(result?.data ?? [], org?.id ?? '', app?.id ?? '')}
      isLoading={isLoading}
      pagination={{
        hasNext: result?.pagination?.hasNext ?? false,
        offset,
        limit: LIMIT,
      }}
    />
  )
}
