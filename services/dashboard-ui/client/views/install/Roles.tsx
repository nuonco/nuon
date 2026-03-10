import { useQuery } from '@tanstack/react-query'
import { EmptyState } from '@/components/common/EmptyState'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { IAMRoles, IAMRolesSkeleton } from '@/components/roles/IAMRoles'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getAppConfig } from '@/lib'

export const Roles = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  const { data: configResult, isLoading } = useQuery({
    queryKey: [
      'app-config',
      org?.id,
      install?.app_id,
      install?.app_config_id,
      'recurse',
    ],
    queryFn: () =>
      getAppConfig({
        orgId: org.id,
        appId: install.app_id,
        appConfigId: install.app_config_id,
        recurse: true,
      }),
    enabled: !!org?.id && !!install?.app_config_id,
  })

  const config = configResult

  return (
    <PageSection isScrollable>
      <PageTitle title={`Roles | ${install?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          {
            path: `/${org?.id}/installs/${install?.id}/roles`,
            text: 'Roles',
          },
        ]}
      />
      <HeadingGroup>
        <Text variant="base" weight="strong">
          IAM roles
        </Text>
        <Text variant="subtext" theme="neutral">
          View the IAM roles that your install uses to access customer AWS
          resources.
        </Text>
      </HeadingGroup>

      {isLoading ? (
        <IAMRolesSkeleton />
      ) : config?.permissions?.aws_iam_roles?.length ? (
        <IAMRoles appConfig={config} />
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
