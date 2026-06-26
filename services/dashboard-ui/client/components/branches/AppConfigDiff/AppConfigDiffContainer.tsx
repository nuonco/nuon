import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { useApp } from '@/hooks/use-app'
import { getAppConfigs, getAppConfigDiff } from '@/lib'
import { AppConfigDiff, extractSections, computeSummary } from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'

interface IAppConfigDiffContainer {
  appConfigId: string
  className?: string
}

export const AppConfigDiffContainer = ({ appConfigId, className }: IAppConfigDiffContainer) => {
  const { org } = useOrg()
  const { app } = useApp()

  const { data: recentConfigs } = useQuery({
    queryKey: ['app-configs', org?.id, app?.id],
    queryFn: () => getAppConfigs({ orgId: org!.id, appId: app!.id, limit: 10 }),
    enabled: !!org?.id && !!app?.id && !!appConfigId,
  })

  const previousConfigs = (recentConfigs || []).filter((c) => c.id !== appConfigId)
  const oldConfigId = previousConfigs[0]?.id

  const { data: diffData, isLoading } = useQuery({
    queryKey: ['app-config-diff', org?.id, app?.id, appConfigId, oldConfigId],
    queryFn: () =>
      getAppConfigDiff({
        orgId: org!.id,
        appId: app!.id,
        configId: appConfigId,
        oldConfigId,
      }),
    enabled: !!org?.id && !!app?.id && !!appConfigId,
    retry: 1,
  })

  const sections = diffData?.diff ? extractSections(diffData.diff) : []
  const summary = sections.length > 0 ? computeSummary(sections) : (diffData?.summary || null)

  return (
    <AppConfigDiff
      sections={sections}
      summary={summary}
      isLoading={isLoading && !diffData}
    />
  )
}
