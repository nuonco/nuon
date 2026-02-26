import { createContext, useEffect } from 'react'
import { useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { getOrg } from '@/lib/ctl-api/orgs'
import { setOrgSession } from '@/lib/cookies'
import type { TOrg } from '@/types'

type OrgContextValue = {
  org: TOrg | null
  isLoading: boolean
  error: unknown
  refresh: () => void
}

export const OrgContext = createContext<OrgContextValue | undefined>(undefined)

export function OrgProvider({ children }: { children: React.ReactNode }) {
  const { orgId } = useParams<{ orgId: string }>()

  const { data: org, isLoading, error, refetch } = useQuery({
    queryKey: ['org', orgId],
    queryFn: () => getOrg({ orgId: orgId! }),
    refetchInterval: 30_000,
    enabled: !!orgId,
  })

  useEffect(() => {
    if (orgId) {
      setOrgSession(orgId)
    }
  }, [orgId])

  return (
    <OrgContext.Provider
      value={{
        org: org ?? null,
        isLoading,
        error,
        refresh: refetch,
      }}
    >
      {children}
    </OrgContext.Provider>
  )
}
