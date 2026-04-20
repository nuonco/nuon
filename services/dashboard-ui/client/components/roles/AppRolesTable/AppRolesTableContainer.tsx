import { useQuery } from '@tanstack/react-query'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getAppConfig, getAppConfigs } from '@/lib'
import { AppRolesTable } from './AppRolesTable'

export const AppRolesTableContainer = () => {
  const { org } = useOrg()
  const { app } = useApp()

  const { data: configs, isLoading: isLoadingConfigs } = useQuery({
    queryKey: ['app-configs', org?.id, app?.id],
    queryFn: () => getAppConfigs({ orgId: org.id, appId: app.id, limit: 1 }),
    enabled: !!org?.id && !!app?.id,
  })

  const appConfigId = configs?.at(0)?.id

  const { data: appConfig, isLoading: isLoadingConfig } = useQuery({
    queryKey: ['app-config', org?.id, app?.id, appConfigId, 'recurse'],
    queryFn: () =>
      getAppConfig({ orgId: org.id, appId: app.id, appConfigId, recurse: true }),
    enabled: !!org?.id && !!app?.id && !!appConfigId,
  })

  const isLoading = isLoadingConfigs || isLoadingConfig

  return (
    <AppRolesTable
      roles={appConfig?.permissions?.aws_iam_roles ?? []}
      isLoading={isLoading}
    />
  )
}
