import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getAppConfig, getAppConfigs } from '@/lib'
import type { TNamedIAMPolicy } from '@/lib/ctl-api/installs/get-install-app-permissions-config'
import { AppRolesTable } from './AppRolesTable'

export const AppRolesTableContainer = () => {
  const { org } = useOrg()
  const { app } = useApp()

  const { data: configs, isLoading: isLoadingConfigs } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-configs', org?.id, app?.id],
    queryFn: () => getAppConfigs({ orgId: org.id, appId: app.id, limit: 1 }),
    enabled: !!org?.id && !!app?.id,
  })

  const appConfigId = configs?.at(0)?.id

  const { data: appConfig, isLoading: isLoadingConfig } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-config', org?.id, app?.id, appConfigId, 'recurse'],
    queryFn: () =>
      getAppConfig({
        orgId: org.id,
        appId: app.id,
        appConfigId,
        recurse: true,
      }),
    enabled: !!org?.id && !!app?.id && !!appConfigId,
  })

  const isLoading = isLoadingConfigs || isLoadingConfig
  const permissions = appConfig?.permissions as
    | { named_policies?: TNamedIAMPolicy[] }
    | undefined

  return (
    <AppRolesTable
      roles={appConfig?.permissions?.aws_iam_roles ?? []}
      namedPolicies={permissions?.named_policies ?? []}
      isLoading={isLoading}
    />
  )
}
