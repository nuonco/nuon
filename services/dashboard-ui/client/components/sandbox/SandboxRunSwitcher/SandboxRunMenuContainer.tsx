import { useEffect, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { type IMenu } from '@/components/common/Menu'
import { useInstall } from '@/hooks/use-install'
import { useScrollToBottom } from '@/hooks/use-on-scroll-to-bottom'
import { useOrg } from '@/hooks/use-org'
import { getInstallSandboxRuns } from '@/lib'
import type { TSandboxRun } from '@/types'
import { SandboxRunMenu } from './SandboxRunMenu'

interface ISandboxRunMenuContainer extends Omit<IMenu, 'children'> {
  activeSandboxRunId: string
}

export const SandboxRunMenuContainer = ({ activeSandboxRunId }: ISandboxRunMenuContainer) => {
  const limit = 8
  const [offset, setOffset] = useState(0)
  const [sandboxRuns, setSandboxRuns] = useState<TSandboxRun[]>([])
  const { install } = useInstall()
  const { org } = useOrg()

  const { data, error, isLoading } = useQuery({
    queryKey: ['sandbox-runs-menu', org?.id, install?.id, offset],
    queryFn: () =>
      getInstallSandboxRuns({
        orgId: org.id,
        installId: install.id,
        limit,
        offset,
      }),
    enabled: !!org?.id && !!install?.id,
  })

  const scrollToBottom = useScrollToBottom({
    onScrollToBottom: () => {
      if (data?.pagination?.hasNext) {
        setOffset((prev) => {
          if (prev === 0) {
            return limit + 1
          } else {
            return prev + limit
          }
        })
      }
    },
  })

  useEffect(() => {
    if (data?.data) {
      setSandboxRuns((prev) => {
        const sandboxRunMap = new Map(
          prev.map((sandboxRun) => [sandboxRun.id, sandboxRun])
        )
        data.data.forEach((sandboxRun) => sandboxRunMap.set(sandboxRun.id, sandboxRun))
        return Array.from(sandboxRunMap.values())
      })
      scrollToBottom.reset()
    }
  }, [data])

  return (
    <SandboxRunMenu
      activeSandboxRunId={activeSandboxRunId}
      sandboxRuns={sandboxRuns}
      isLoading={isLoading}
      hasError={!!error}
      orgId={org.id}
      installId={install.id}
      scrollRef={scrollToBottom.elementRef}
      limit={limit}
    />
  )
}
