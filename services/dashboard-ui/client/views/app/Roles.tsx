import { useQuery } from '@tanstack/react-query'
import { IAMRoles, IAMRolesSkeleton } from '@/components/roles/IAMRoles'
import { EmptyState } from '@/components/common/EmptyState/EmptyState'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getAppConfig, getAppConfigs } from '@/lib'

export const Roles = () => {
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
    <PageSection isScrollable>
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          { path: `/${org?.id}/apps/${app?.id}/roles`, text: 'Roles' },
        ]}
      />
      <HeadingGroup>
        <Text variant="base" weight="strong">
          IAM roles
        </Text>
        <Text variant="subtext" theme="neutral">
          View the IAM roles that your app uses to access customer AWS resources.
        </Text>
      </HeadingGroup>

      {isLoading ? (
        <IAMRolesSkeleton />
      ) : appConfig?.permissions?.aws_iam_roles?.length ? (
        <IAMRoles appConfig={appConfig} />
      ) : (
        <EmptyState
          variant="table"
          emptyTitle="No roles found"
          emptyMessage="You don't have any roles assigned yet. Contact your administrator to get access to roles."
        />
      )}
    </PageSection>
  )
}
