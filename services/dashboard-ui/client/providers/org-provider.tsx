import { createContext, useEffect } from 'react'
import { useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { getOrg } from '@/lib/ctl-api/orgs'
import { clearOrgSession, setOrgSession } from '@/lib/cookies'
import { ProviderError } from '@/components/layout/ProviderError'
import { ProviderLoading } from '@/components/layout/ProviderLoading'
import type { TAPIError, TOrg } from '@/types'

type OrgContextValue = {
  org: TOrg
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

  // If the org doesn't exist (404/403), clear the stale session cookie
  // and redirect to / so the BFF can resolve a valid org via GetOrgs.
  useEffect(() => {
    if (!error) return
    const status = (error as TAPIError)?.status
    if (status === 404 || status === 403) {
      clearOrgSession()
      window.location.href = '/'
    }
  }, [error])

  if (error && !org) return <ProviderError error={error} />

  if (isLoading || !org) return <ProviderLoading />

  return (
    <OrgContext.Provider value={{ org, refresh: refetch }}>
      {children}
    </OrgContext.Provider>
  )
}
