import { EmptyState } from '@/components/common/EmptyState'
import { IAMRoles, IAMRolesSkeleton } from '@/components/roles/IAMRoles'
import { getAppConfigs, getAppConfig } from '@/lib'

export const AppRoles = async ({
  appId,
  orgId,
}: {
  appId: string
  orgId: string
}) => {
  const { data: configs, error: configsError } = await getAppConfigs({
    appId,
    orgId,
  })

  if (configsError) {
    return <AppRolesError />
  }

  const { data: config, error } = await getAppConfig({
    appConfigId: configs?.at(0)?.id,
    appId,
    orgId,
    recurse: true,
  })

  return error ? (
    <AppRolesError />
  ) : config?.permissions?.aws_iam_roles?.length ? (
    <IAMRoles appConfig={config} />
  ) : (
    <AppRolesError
      {...{
        title: 'No roles found',
        message:
          "You don't have any roles assigned yet. Contact your administrator to get access to roles.",
      }}
    />
  )
}

export const AppRolesSkeleton = IAMRolesSkeleton

export const AppRolesError = ({
  title = 'Unable to load roles',
  message = 'We encountered an issue loading your roles. Please try refreshing the page or contact support if the problem persists.',
}: {
  title?: string
  message?: string
}) => {
  return (
    <EmptyState variant="table" emptyMessage={message} emptyTitle={title} />
  )
}
