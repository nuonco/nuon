import {
  createContext,
  useContext,
  useEffect,
  useMemo,
  type ReactNode,
} from 'react'
import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router'
import { getOrg } from '@/lib'
import { clearOrgSession, setOrgSession } from '@/lib/cookies'
import type { TAPIError, TOrg } from '@/types'

interface IOrgContext {
  org?: TOrg
  orgId?: string
  isLoading: boolean
  error: unknown
  refresh: () => void
}

const OrgContext = createContext<IOrgContext | null>(null)

export const OrgProvider = ({ children }: { children: ReactNode }) => {
  const { orgId } = useParams<{ orgId: string }>()
  const {
    data: org,
    isLoading,
    error,
    refetch,
  } = useQuery({
    queryKey: ['org', orgId],
    queryFn: () => getOrg({ orgId: orgId! }),
    enabled: !!orgId,
    refetchInterval: 30_000,
    retry: false,
  })

  useEffect(() => {
    if (org && orgId) setOrgSession(orgId)
  }, [org, orgId])

  useEffect(() => {
    const status = (error as TAPIError | undefined)?.status
    if (status !== 403 && status !== 404) return
    clearOrgSession()
    window.location.assign('/')
  }, [error])

  const value = useMemo(
    () => ({
      org,
      orgId,
      isLoading,
      error,
      refresh: () => void refetch(),
    }),
    [error, isLoading, org, orgId, refetch]
  )

  return <OrgContext.Provider value={value}>{children}</OrgContext.Provider>
}

export const useOrg = () => {
  const context = useContext(OrgContext)
  if (!context) {
    throw new Error('useOrg must be used within OrgProvider')
  }
  return context
}
