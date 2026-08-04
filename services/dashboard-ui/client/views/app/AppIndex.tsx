import { useEffect } from 'react'
import { useNavigate } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { PageContent } from '@/components/layout/PageContent'
import { ProviderLoading } from '@/components/layout/ProviderLoading'
import { useApp } from '@/hooks/use-app'
import { useNewAppIA } from '@/hooks/use-new-app-ia'
import { useOrg } from '@/hooks/use-org'
import { getAppBranches } from '@/lib'
import { Overview } from './Overview'
import { Branches } from './branches/Branches'

const BranchPicker = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const navigate = useNavigate()

  const { data: result, isLoading } = useQuery({
    queryKey: ['app-branches-source', org?.id, app?.id],
    queryFn: () =>
      getAppBranches({ orgId: org!.id!, appId: app!.id!, limit: 50, offset: 0 }),
    enabled: !!org?.id && !!app?.id,
  })

  const branches = result?.data ?? []
  const singleBranchId = branches.length === 1 ? branches[0].id : undefined

  useEffect(() => {
    if (isLoading || !singleBranchId) return
    navigate(`/${org?.id}/apps/${app?.id}/branches/${singleBranchId}`, {
      replace: true,
    })
  }, [isLoading, singleBranchId, navigate, org?.id, app?.id])

  if (isLoading || singleBranchId) {
    return (
      <PageContent className="border-t">
        <ProviderLoading />
      </PageContent>
    )
  }

  return (
    <PageContent className="border-t">
      <Branches />
    </PageContent>
  )
}

export const AppIndex = () => {
  const hasNewAppIA = useNewAppIA()

  return hasNewAppIA ? <BranchPicker /> : <Overview />
}
