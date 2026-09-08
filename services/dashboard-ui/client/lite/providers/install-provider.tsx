import {
  createContext,
  useContext,
  useMemo,
  type ReactNode,
} from 'react'
import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router'
import { getInstall } from '@/lib'
import type { TInstall } from '@/types'
import { useOrg } from './org-provider'

interface IInstallContext {
  install?: TInstall
  installId?: string
  isLoading: boolean
  error: unknown
  refresh: () => void
}

const InstallContext = createContext<IInstallContext | null>(null)

export const InstallProvider = ({ children }: { children: ReactNode }) => {
  const { orgId } = useOrg()
  const { installId } = useParams<{ installId: string }>()
  const {
    data: install,
    isLoading,
    error,
    refetch,
  } = useQuery({
    queryKey: ['install', orgId, installId],
    queryFn: () => getInstall({ orgId: orgId!, installId: installId! }),
    enabled: !!orgId && !!installId,
    refetchInterval: 20_000,
  })

  const value = useMemo(
    () => ({
      install,
      installId,
      isLoading,
      error,
      refresh: () => void refetch(),
    }),
    [error, install, installId, isLoading, refetch]
  )

  return (
    <InstallContext.Provider value={value}>{children}</InstallContext.Provider>
  )
}

export const useInstall = () => {
  const context = useContext(InstallContext)
  if (!context) {
    throw new Error('useInstall must be used within InstallProvider')
  }
  return context
}
