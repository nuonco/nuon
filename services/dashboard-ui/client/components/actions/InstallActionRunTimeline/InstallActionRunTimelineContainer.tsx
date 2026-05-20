import { useMemo } from 'react'
import { useSearchParams } from 'react-router'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useResourceSSE } from '@/hooks/use-resource-sse'
import { getInstallAction } from '@/lib'
import { InstallActionRunTimeline } from './InstallActionRunTimeline'

const LIMIT = 10

interface IInstallActionRunTimelineContainer {
  actionId: string
  actionName: string
  pollInterval?: number
  shouldPoll?: boolean
}

export const InstallActionRunTimelineContainer = ({
  actionId,
  actionName,
  pollInterval = 20000,
  shouldPoll = false,
}: IInstallActionRunTimelineContainer) => {
  const { install } = useInstall()
  const { org } = useOrg()
  const queryClient = useQueryClient()
  const [searchParams] = useSearchParams()
  const offset = Number(searchParams.get('offset') ?? 0)

  const queryKey = useMemo(
    () => ['install-action', org?.id, install?.id, actionId, offset],
    [org?.id, install?.id, actionId, offset]
  )

  const sseUrl =
    org?.id && install?.id && actionId
      ? `/api/orgs/${org.id}/installs/${install.id}/actions/${actionId}/runs/sse?limit=${LIMIT}&offset=${offset}`
      : undefined

  const listeners = useMemo(
    () => ({
      'action-runs': (event: MessageEvent) => {
        try {
          const data = JSON.parse(event.data)
          queryClient.setQueryData(queryKey, data?.data)
        } catch {}
      },
    }),
    [queryKey, queryClient]
  )

  const { connected: sseConnected } = useResourceSSE({
    url: sseUrl,
    enabled: shouldPoll,
    listeners,
  })

  const { data: action } = useQuery({
    queryKey,
    queryFn: () =>
      getInstallAction({
        orgId: org.id,
        installId: install.id,
        actionId,
        limit: LIMIT,
        offset,
      }),
    refetchOnMount: 'always',
    refetchInterval: shouldPoll && !sseConnected ? pollInterval : false,
    enabled: !!org?.id && !!install?.id && !!actionId,
  })

  const runs = action?.runs ?? []
  const basePath = `/${org.id}/installs/${install.id}`

  return (
    <InstallActionRunTimeline
      actionId={actionId}
      actionName={actionName}
      runs={runs}
      basePath={basePath}
      pagination={{ hasNext: runs.length >= LIMIT, offset, limit: LIMIT }}
    />
  )
}
