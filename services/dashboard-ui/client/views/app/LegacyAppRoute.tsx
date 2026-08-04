import { useEffect, type ReactNode } from 'react'
import { useNavigate, useParams, type Params } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { PageContent } from '@/components/layout/PageContent'
import { ProviderLoading } from '@/components/layout/ProviderLoading'
import { useApp } from '@/hooks/use-app'
import { useNewAppIA } from '@/hooks/use-new-app-ia'
import { useOrg } from '@/hooks/use-org'
import { getAppBranches } from '@/lib'

export const LegacyAppRoute = ({
  subPath,
  children,
}: {
  subPath?: (params: Params) => string
  children: ReactNode
}) => {
  const hasNewAppIA = useNewAppIA()
  const { org } = useOrg()
  const { app } = useApp()
  const params = useParams()
  const navigate = useNavigate()

  const { data: result, isLoading } = useQuery({
    queryKey: ['app-branches-source', org?.id, app?.id],
    queryFn: () =>
      getAppBranches({ orgId: org!.id!, appId: app!.id!, limit: 50, offset: 0 }),
    enabled: hasNewAppIA && !!org?.id && !!app?.id,
  })

  const targetBranchId = result?.data?.[0]?.id

  useEffect(() => {
    if (!hasNewAppIA || isLoading || !org?.id || !app?.id) return

    if (!targetBranchId) {
      navigate(`/${org.id}/apps/${app.id}`, { replace: true })
      return
    }

    const suffix = subPath ? `/${subPath(params)}` : ''
    navigate(
      `/${org.id}/apps/${app.id}/branches/${targetBranchId}${suffix}`,
      { replace: true }
    )
  }, [
    hasNewAppIA,
    isLoading,
    targetBranchId,
    org?.id,
    app?.id,
    navigate,
    subPath,
    params,
  ])

  if (!hasNewAppIA) return <>{children}</>

  return (
    <PageContent className="border-t">
      <ProviderLoading />
    </PageContent>
  )
}
