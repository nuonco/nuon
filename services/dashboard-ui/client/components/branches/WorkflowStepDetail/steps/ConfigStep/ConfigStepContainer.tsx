import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { useApp } from '@/hooks/use-app'
import { getAppConfigs, getAppConfigDiff } from '@/lib'
import { ConfigStep } from './ConfigStep'
import { extractSections, computeSummary } from './lib'

interface IConfigStepContainer {
  metadata: Record<string, any>
  status?: string
}

export const ConfigStepContainer = ({ metadata, status }: IConfigStepContainer) => {
  const { org } = useOrg()
  const { app } = useApp()
  const appConfigId = metadata.app_config_id as string | undefined

  const { data: recentConfigs } = useQuery({
    queryKey: ['app-configs', org?.id, app?.id],
    queryFn: () => getAppConfigs({ orgId: org!.id, appId: app!.id, limit: 10 }),
    enabled: !!org?.id && !!app?.id && !!appConfigId,
  })

  const previousConfigs = (recentConfigs || []).filter((c: any) => c.id !== appConfigId)
  const oldConfigId = previousConfigs[0]?.id

  const { data: diffData, isError: diffError } = useQuery({
    queryKey: ['app-config-diff', org?.id, app?.id, appConfigId, oldConfigId],
    queryFn: () =>
      getAppConfigDiff({
        orgId: org!.id,
        appId: app!.id,
        configId: appConfigId!,
        oldConfigId,
      }),
    enabled: !!org?.id && !!app?.id && !!appConfigId,
    retry: 1,
  })

  const sections = diffData?.diff ? extractSections(diffData.diff) : []
  const summary = sections.length > 0 ? computeSummary(sections) : (diffData?.summary || null)

  return (
    <ConfigStep
      appConfigId={appConfigId}
      status={status}
      sections={sections}
      summary={summary}
      diffResolved={diffError || !!diffData}
      metadata={metadata}
    />
  )
}
