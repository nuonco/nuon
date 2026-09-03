import { useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getAppReleases } from '@/lib'
import { ReleasesTable } from './ReleasesTable'

const LIMIT = 20

export const ReleasesTableContainer = () => {
  const [searchParams] = useSearchParams()
  const { app } = useApp()
  const { org } = useOrg()
  const offset = Number(searchParams.get('offset') ?? 0)
  const { data: result, isLoading } = useQuery({
    queryKey: ['app-releases', org?.id, app?.id, offset],
    queryFn: () =>
      getAppReleases({
        appId: app!.id,
        limit: LIMIT,
        offset,
        orgId: org!.id,
      }),
    enabled: !!org?.id && !!app?.id,
    placeholderData: keepPreviousData,
  })

  return (
    <ReleasesTable
      appId={app?.id ?? ''}
      data={result?.data ?? []}
      isLoading={isLoading}
      orgId={org?.id ?? ''}
      pagination={{
        hasNext: result?.pagination?.hasNext ?? false,
        limit: LIMIT,
        offset,
      }}
    />
  )
}
