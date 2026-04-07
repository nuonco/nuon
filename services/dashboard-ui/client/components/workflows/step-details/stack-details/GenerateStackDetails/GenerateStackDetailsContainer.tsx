import { useQuery } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getAppConfig } from '@/lib'
import type { TAppConfig } from '@/types'
import { GenerateStackDetails } from './GenerateStackDetails'

export const GenerateStackDetailsContainer = () => {
  const { install } = useInstall()
  const { org } = useOrg()
  const { data: appConfig, isLoading } = useQuery<TAppConfig>({
    queryKey: ['app-config', org?.id, install?.app_id, install?.app_config_id],
    queryFn: () =>
      getAppConfig({
        orgId: org.id,
        appId: install.app_id,
        appConfigId: install.app_config_id,
        recurse: true,
      }),
    enabled: !!org?.id && !!install?.app_id && !!install?.app_config_id,
  })

  return (
    <GenerateStackDetails appConfig={appConfig} isLoading={isLoading} />
  )
}
