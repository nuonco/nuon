import { useMemo } from 'react'
import { useSearchParams } from 'react-router'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useResourceSSE } from '@/hooks/use-resource-sse'
import { getInstallWorkflows } from '@/lib'
import { WorkflowTimeline, WorkflowTimelineSkeleton } from './WorkflowTimeline'

export { WorkflowTimelineSkeleton }

const LIMIT = 10

interface IWorkflowTimelineContainer {
  installId: string
  pollInterval?: number
  shouldPoll?: boolean
  type?: string
  planonly?: boolean
  search?: string
}

export const WorkflowTimelineContainer = ({
  installId,
  shouldPoll = false,
  pollInterval = 20000,
  planonly = true,
  type = '',
  search = '',
}: IWorkflowTimelineContainer) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const queryClient = useQueryClient()
  const [searchParams] = useSearchParams()
  const offset = Number(searchParams.get('offset') ?? 0)

  const queryKey = useMemo(
    () => ['install-workflows', org?.id, installId, offset, planonly, type, search],
    [org?.id, installId, offset, planonly, type, search]
  )

  const sseUrl =
    org?.id && installId
      ? `/api/orgs/${org.id}/installs/${installId}/workflows/sse?limit=${LIMIT}&offset=${offset}&planonly=${planonly}&type=${type}&search=${encodeURIComponent(search)}`
      : undefined

  const activeWorkflowsQueryKey = useMemo(
    () => ['install-active-workflows', org?.id, install?.id],
    [org?.id, install?.id]
  )

  const listeners = useMemo(
    () => ({
      workflows: (event: MessageEvent) => {
        try {
          const data = JSON.parse(event.data)
          queryClient.setQueryData(queryKey, data)
        } catch {}
      },
      'active-workflows': (event: MessageEvent) => {
        try {
          const data = JSON.parse(event.data)
          queryClient.setQueryData(activeWorkflowsQueryKey, { data })
        } catch {}
      },
    }),
    [queryKey, activeWorkflowsQueryKey, queryClient]
  )

  const { connected: sseConnected } = useResourceSSE({
    url: sseUrl,
    enabled: shouldPoll,
    listeners,
  })

  const { data: result, isLoading } = useQuery({
    queryKey,
    queryFn: () =>
      getInstallWorkflows({
        orgId: org.id,
        installId,
        limit: LIMIT,
        offset,
        planonly,
        type,
        search,
      }),
    refetchOnMount: 'always',
    refetchInterval: shouldPoll && !sseConnected ? pollInterval : false,
    enabled: !!org?.id && !!installId,
  })

  const workflows = result?.data ?? []
  const pagination = result?.pagination
    ? { hasNext: result.pagination.hasNext, offset, limit: LIMIT }
    : { hasNext: false, offset, limit: LIMIT }

  return (
    <WorkflowTimeline
      workflows={workflows}
      pagination={pagination}
      orgId={org?.id}
      installId={installId}
      install={install}
      isLoading={isLoading}
    />
  )
}
