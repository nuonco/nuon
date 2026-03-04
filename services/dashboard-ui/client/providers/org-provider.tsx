import { createContext, useEffect } from 'react'
import { useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { getOrg } from '@/lib/ctl-api/orgs'
import { setOrgSession } from '@/lib/cookies'
import { Loading } from '@/components/common/Loading'
import type { TOrg } from '@/types'

type OrgContextValue = {
  org: TOrg
  refresh: () => void
}

export const OrgContext = createContext<OrgContextValue | undefined>(undefined)

export function OrgProvider({ children }: { children: React.ReactNode }) {
  const { orgId } = useParams<{ orgId: string }>()

  const { data: org, isLoading, refetch } = useQuery({
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

  if (isLoading || !org) return <Loading />

  return (
    <OrgContext.Provider value={{ org, refresh: refetch }}>
      {children}
    </OrgContext.Provider>
  )
}
