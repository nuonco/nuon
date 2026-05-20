import { useMemo } from 'react'
import { useSearchParams } from 'react-router'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useResourceSSE } from '@/hooks/use-resource-sse'
import { getComponentDeploys } from '@/lib'
import { DeployTimeline } from './DeployTimeline'

const LIMIT = 10

interface IDeployTimelineContainer {
  componentName: string
  componentId: string
  pollInterval?: number
  shouldPoll?: boolean
}

export const DeployTimelineContainer = ({
  componentName,
  componentId,
  pollInterval = 20000,
  shouldPoll = false,
}: IDeployTimelineContainer) => {
  const { install } = useInstall()
  const { org } = useOrg()
  const queryClient = useQueryClient()
  const [searchParams] = useSearchParams()
  const offset = Number(searchParams.get('offset') ?? 0)

  const queryKey = useMemo(
    () => ['component-deploys', org?.id, install?.id, componentId, offset],
    [org?.id, install?.id, componentId, offset]
  )

  const sseUrl =
    org?.id && install?.id && componentId
      ? `/api/orgs/${org.id}/installs/${install.id}/components/${componentId}/deploys/sse?limit=${LIMIT}&offset=${offset}`
      : undefined

  const listeners = useMemo(
    () => ({
      deploys: (event: MessageEvent) => {
        try {
          const data = JSON.parse(event.data)
          queryClient.setQueryData(queryKey, data)
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

  const { data: result, isLoading, error } = useQuery({
    queryKey,
    queryFn: () =>
      getComponentDeploys({
        orgId: org.id,
        installId: install.id,
        componentId,
        limit: LIMIT,
        offset,
      }),
    refetchOnMount: 'always',
    refetchInterval: shouldPoll && !sseConnected ? pollInterval : false,
    enabled: !!org?.id && !!install?.id && !!componentId,
  })

  const deploys = result?.data ?? []
  const pagination = result?.pagination
    ? { hasNext: result.pagination.hasNext, offset, limit: LIMIT }
    : { hasNext: false, offset, limit: LIMIT }

  return (
    <DeployTimeline
      deploys={deploys}
      pagination={pagination}
      orgId={org?.id}
      installId={install?.id}
      componentId={componentId}
      componentName={componentName}
      isLoading={isLoading}
      error={error}
    />
  )
}
