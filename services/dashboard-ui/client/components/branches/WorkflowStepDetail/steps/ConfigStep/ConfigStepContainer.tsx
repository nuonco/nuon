import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getAppConfigDiff } from '@/lib'
import { extractSections } from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'
import { ConfigStep } from './ConfigStep'

interface IConfigStepContainer {
  metadata: Record<string, any>
  status?: string
}

export const ConfigStepContainer = ({
  metadata,
  status,
}: IConfigStepContainer) => {
  const { org } = useOrg()
  const { app } = useApp()
  const appConfigId = metadata.app_config_id as string | undefined

  const { data, isLoading, isError } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-config-snapshot', org?.id, app?.id, appConfigId],
    queryFn: () =>
      getAppConfigDiff({
        orgId: org!.id,
        appId: app!.id,
        configId: appConfigId!,
      }),
    enabled: !!org?.id && !!app?.id && !!appConfigId,
  })

  const sections = extractSections(data?.diff)

  return (
    <ConfigStep
      appConfigId={appConfigId}
      status={status}
      sections={sections}
      isLoading={isLoading && !data}
      isError={isError}
    />
  )
}
