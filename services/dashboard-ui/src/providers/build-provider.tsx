'use client'

import { createContext, type ReactNode } from 'react'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import { useOrg } from '@/hooks/use-org'
import type { TBuild } from '@/types'

type BuildContextValue = {
  build: TBuild | null
  isLoading: boolean
  error: any
}

export const BuildContext = createContext<BuildContextValue | undefined>(
  undefined
)

export function BuildProvider({
  children,
  initBuild,
  pollInterval = 10000,
  shouldPoll = true,
}: {
  children: ReactNode
  initBuild: TBuild
} & IPollingProps) {
  const { org } = useOrg()
  const {
    data: build,
    error,
    isLoading,
  } = usePolling<TBuild>({
    dependencies: [initBuild],
    initData: initBuild,
    path: `/api/orgs/${org.id}/components/${initBuild?.component_id}/builds/${initBuild.id}`,
    pollInterval,
    shouldPoll,
  })

  return (
    <BuildContext.Provider
      value={{
        build,
        isLoading,
        error,
      }}
    >
      {children}
    </BuildContext.Provider>
  )
}
