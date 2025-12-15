import { EmptyState } from '@/components/common/EmptyState'
import { IAMRoles, IAMRolesSkeleton } from '@/components/roles/IAMRoles'
import { getAppConfig } from '@/lib'

export const InstallRoles = async ({
  appConfigId,
  appId,
  orgId,
}: {
  appConfigId: string
  appId: string
  orgId: string
}) => {

  const { data: config, error } = await getAppConfig({
    appConfigId,
    appId,
    orgId,
    recurse: true,
  })

  return error ? (
    <InstallRolesError />
  ) : config?.permissions?.aws_iam_roles?.length ? (
    <IAMRoles appConfig={config} />
  ) : (
    <InstallRolesError
      {...{
        title: 'No roles found',
        message:
          "You don't have any roles assigned yet. Contact your administrator to get access to roles.",
      }}
    />
  )
}

export const InstallRolesSkeleton = IAMRolesSkeleton

export const InstallRolesError = ({
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
