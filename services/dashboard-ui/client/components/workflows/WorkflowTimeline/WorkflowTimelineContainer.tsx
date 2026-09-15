import { useMemo } from 'react'
import { useSearchParams } from 'react-router'
import { useQueryClient } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useSSETimelineQuery } from '@/lib/sse/use-sse-timeline-query'
import { useRefreshErrorToast } from '@/hooks/use-refresh-error-toast'
import { getInstallWorkflows } from '@/lib'
import { createSSEQueryListener } from '@/lib/sse-listeners'
import { WorkflowTimeline } from './WorkflowTimeline'

const LIMIT = 10

interface IWorkflowTimelineContainer {
  installId: string
  pollInterval?: number
  shouldPoll?: boolean
  type?: string
  planonly?: boolean
  search?: string
  status?: string
  createdAtGte?: string
  isFiltered?: boolean
}

export const WorkflowTimelineContainer = ({
  installId,
  shouldPoll = false,
  pollInterval = 20000,
  planonly = true,
  type = '',
  search = '',
  status = '',
  createdAtGte,
  isFiltered = false,
}: IWorkflowTimelineContainer) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const queryClient = useQueryClient()
  const [searchParams] = useSearchParams()
  const offset = Number(searchParams.get('offset') ?? 0)

  const extraListeners = useMemo(
    () => ({
      'active-workflows': createSSEQueryListener(
        queryClient,
        ['install-active-workflows', org?.id, install?.id],
        { transform: (data) => ({ data }) }
      ),
    }),
    [queryClient, org?.id, install?.id]
  )

  const onRefreshError = useRefreshErrorToast()
  const sseUrl = useMemo(() => {
    if (!org?.id || !installId) return undefined

    const params = new URLSearchParams({
      limit: String(LIMIT),
      offset: String(offset),
      planonly: String(planonly),
    })
    if (type) params.set('type', type)
    if (status) params.set('status', status)
    if (search) params.set('search', search)
    if (createdAtGte) params.set('created_at_gte', createdAtGte)
    return `/api/orgs/${org.id}/installs/${installId}/workflows/sse?${params}`
  }, [createdAtGte, installId, offset, org?.id, planonly, search, status, type])

  const { data: result, isLoading } = useSSETimelineQuery({
    sseUrl,
    queryKey: [
      'install-workflows',
      org?.id,
      installId,
      offset,
      planonly,
      type,
      search,
      status,
      createdAtGte,
    ],
    queryFn: () =>
      getInstallWorkflows({
        orgId: org.id,
        installId,
        limit: LIMIT,
        offset,
        planonly,
        type,
        search,
        status,
        created_at_gte: createdAtGte,
      }),
    enabled: !!org?.id && !!installId,
    shouldPoll,
    pollInterval,
    eventName: 'workflows',
    onError: onRefreshError,
    extraListeners,
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
      isFiltered={isFiltered}
    />
  )
}
