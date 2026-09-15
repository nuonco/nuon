import { useEffect } from 'react'
import { useNavigate } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { PageContent } from '@/components/layout/PageContent'
import { ProviderLoading } from '@/components/layout/ProviderLoading'
import { useApp } from '@/hooks/use-app'
import { useNewAppIA } from '@/hooks/use-new-app-ia'
import { useOrg } from '@/hooks/use-org'
import { useSimpleIA } from '@/hooks/use-simple-ia'
import { getAppBranches } from '@/lib'
import { Overview } from './Overview'
import { Branches } from './branches/Branches'

const BranchPicker = ({
  resolveFirstBranch = false,
}: {
  resolveFirstBranch?: boolean
}) => {
  const { org } = useOrg()
  const { app } = useApp()
  const navigate = useNavigate()

  const { data: result, isLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-branches-source', org?.id, app?.id],
    queryFn: () =>
      getAppBranches({
        orgId: org!.id!,
        appId: app!.id!,
        limit: 50,
        offset: 0,
      }),
    enabled: !!org?.id && !!app?.id,
  })

  const branches = result?.data ?? []
  const singleBranchId = branches.length === 1 ? branches[0].id : undefined
  const targetBranchId = resolveFirstBranch ? branches[0]?.id : singleBranchId

  useEffect(() => {
    if (isLoading || !resolveFirstBranch || targetBranchId) return
    navigate(`/${org?.id}/apps/setup?appId=${app?.id}`, { replace: true })
  }, [
    isLoading,
    resolveFirstBranch,
    targetBranchId,
    navigate,
    org?.id,
    app?.id,
  ])

  useEffect(() => {
    if (isLoading || !targetBranchId) return
    navigate(`/${org?.id}/apps/${app?.id}/branches/${targetBranchId}`, {
      replace: true,
    })
  }, [isLoading, targetBranchId, navigate, org?.id, app?.id])

  if (isLoading || resolveFirstBranch || singleBranchId) {
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
  const hasSimpleIA = useSimpleIA()

  if (hasSimpleIA) return <BranchPicker resolveFirstBranch />
  return hasNewAppIA ? <BranchPicker /> : <Overview />
}
