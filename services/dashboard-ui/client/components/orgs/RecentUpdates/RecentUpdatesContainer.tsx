import { useMemo } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useInstallNested } from '@/hooks/use-install-path'
import { useOrg } from '@/hooks/use-org'
import { useRefreshErrorToast } from '@/hooks/use-refresh-error-toast'
import { useResourceSSE } from '@/lib/sse/use-resource-sse'
import { createSSEQueryListener } from '@/lib/sse-listeners'
import {
  toRecentUpdateItem,
  type TRecentUpdatesPayload,
} from './map-recent-updates'
import { RecentUpdates } from './RecentUpdates'

const recentUpdatesKey = (orgId?: string) => ['recent-updates', orgId] as const

export const RecentUpdatesContainer = () => {
  const { org } = useOrg()
  const nestedInstalls = useInstallNested()
  const queryClient = useQueryClient()
  const onRefreshError = useRefreshErrorToast()
  const orgId = org?.id

  const listeners = useMemo(
    () => ({
      'recent-updates': createSSEQueryListener<TRecentUpdatesPayload>(
        queryClient,
        recentUpdatesKey(orgId)
      ),
    }),
    [queryClient, orgId]
  )

  useResourceSSE({
    url: orgId ? `/api/orgs/${orgId}/recent-updates/sse` : undefined,
    enabled: !!orgId,
    listeners,
    onError: onRefreshError,
  })

  const { data } = useQuery({
    queryKey: recentUpdatesKey(orgId),
    queryFn: (): Promise<TRecentUpdatesPayload> =>
      Promise.resolve({ updates: [] }),
    enabled: false,
  })

  const items = useMemo(
    () =>
      (data?.updates ?? []).flatMap((update) => {
        const item = toRecentUpdateItem({
          ...update,
          orgId: orgId ?? '',
          nestedInstalls,
        })
        return item ? [item] : []
      }),
    [data, orgId, nestedInstalls]
  )

  return <RecentUpdates items={items} isLoading={data === undefined} />
}
