import {
  createContext,
  useContext,
  useMemo,
  type ReactNode,
} from 'react'
import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router'
import { getApp } from '@/lib'
import type { TApp } from '@/types'
import { useOrg } from './org-provider'

interface IAppContext {
  app?: TApp
  appId?: string
  loading: boolean
  error: unknown
  refresh: () => void
}

const AppContext = createContext<IAppContext | null>(null)

export const AppProvider = ({ children }: { children: ReactNode }) => {
  const { orgId } = useOrg()
  const { appId } = useParams<{ appId: string }>()
  const {
    data: app,
    isLoading,
    error,
    refetch,
  } = useQuery({
    queryKey: ['app', orgId, appId],
    queryFn: () => getApp({ orgId: orgId!, appId: appId! }),
    enabled: !!orgId && !!appId,
    refetchInterval: 20_000,
  })

  const value = useMemo(
    () => ({
      app,
      appId,
      loading: isLoading,
      error,
      refresh: () => void refetch(),
    }),
    [app, appId, error, isLoading, refetch]
  )

  return <AppContext.Provider value={value}>{children}</AppContext.Provider>
}

export const useApp = () => {
  const context = useContext(AppContext)
  if (!context) {
    throw new Error('useApp must be used within AppProvider')
  }
  return context
}
