import { createContext, useEffect, useRef, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { getComponentBuild } from '@/lib'
import { Toast } from '@/components/surfaces/Toast'
import { Badge } from '@/components/common/Badge'
import { ProviderError } from '@/components/layout/ProviderError'
import { ProviderLoading } from '@/components/layout/ProviderLoading'
import type { TAPIError, TBuild } from '@/types'

type BuildContextValue = {
  build: TBuild
}

export const BuildContext = createContext<BuildContextValue | undefined>(
  undefined
)

export function BuildProvider({
  children,
  buildId,
  componentId,
  componentName,
  pollInterval = 10000,
  shouldPoll = true,
  watchBuild = false,
}: {
  children: ReactNode
  buildId: string
  componentId: string
  componentName?: string
  pollInterval?: number
  shouldPoll?: boolean
  watchBuild?: boolean
}) {
  const { org } = useOrg()
  const { addToast } = useToast()
  const { data: build, isLoading, error } = useQuery({
    queryKey: ['build', org.id!, componentId, buildId],
    queryFn: () => getComponentBuild({ orgId: org.id!, componentId, buildId }),
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org.id && !!componentId && !!buildId,
  })

  const toastFiredRef = useRef(false)

  useEffect(() => {
    if (!watchBuild || toastFiredRef.current) return
    const currentStatus = build?.status_v2?.status
    if (currentStatus !== 'active' && currentStatus !== 'error') return
    toastFiredRef.current = true

    const nameLabel = componentName ? (
      <Badge variant="code" size="md">{componentName}</Badge>
    ) : null

    if (currentStatus === 'active') {
      addToast(
        <Toast heading={<span className="inline-flex items-center gap-1.5">{nameLabel} build succeeded</span>} theme="success" />
      )
    } else if (currentStatus === 'error') {
      addToast(
        <Toast heading={<span className="inline-flex items-center gap-1.5">{nameLabel} build failed</span>} theme="error" />
      )
    }
  }, [build?.status_v2?.status])

  useEffect(() => {
    if (error && build) {
      addToast(
        <Toast heading="Failed to refresh data" theme="warn">
          {(error as TAPIError)?.error ?? 'Connection issue'}
        </Toast>
      )
    }
  }, [error])

  if (error && !build) return <ProviderError error={error} />

  if (isLoading || !build) return <ProviderLoading />

  return (
    <BuildContext.Provider value={{ build }}>
      {children}
    </BuildContext.Provider>
  )
}
