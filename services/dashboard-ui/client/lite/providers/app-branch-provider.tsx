import {
  createContext,
  useContext,
  useEffect,
  useMemo,
  type ReactNode,
} from 'react'
import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router'
import { getAppBranch } from '@/lib'
import type { TAppBranch } from '@/types'
import { setLastAppBranch } from '../utils/app-branch-session'
import { useApp } from './app-provider'
import { useOrg } from './org-provider'

interface IAppBranchContext {
  branch?: TAppBranch
  branchId?: string
  loading: boolean
  error: unknown
  refresh: () => void
}

const AppBranchContext = createContext<IAppBranchContext | null>(null)

export const AppBranchProvider = ({ children }: { children: ReactNode }) => {
  const { orgId } = useOrg()
  const { appId } = useApp()
  const { branchId } = useParams<{ branchId: string }>()
  const {
    data: branch,
    isLoading,
    error,
    refetch,
  } = useQuery({
    queryKey: ['app-branch', orgId, appId, branchId],
    queryFn: () =>
      getAppBranch({
        orgId: orgId!,
        appId: appId!,
        branchId: branchId!,
      }),
    enabled: !!orgId && !!appId && !!branchId,
    refetchInterval: 20_000,
  })

  useEffect(() => {
    if (orgId && appId && branch?.id) setLastAppBranch(orgId, appId, branch.id)
  }, [appId, branch?.id, orgId])

  const value = useMemo(
    () => ({
      branch,
      branchId,
      loading: isLoading,
      error,
      refresh: () => void refetch(),
    }),
    [branch, branchId, error, isLoading, refetch]
  )

  return (
    <AppBranchContext.Provider value={value}>
      {children}
    </AppBranchContext.Provider>
  )
}

export const useAppBranch = () => {
  const context = useContext(AppBranchContext)
  if (!context) {
    throw new Error('useAppBranch must be used within AppBranchProvider')
  }
  return context
}
