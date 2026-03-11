import { useQuery } from '@tanstack/react-query'
import { EmptyState } from '@/components/common/EmptyState'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { InstallIAMRoles, IAMRolesSkeleton } from '@/components/roles/IAMRoles'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallAppPermissionsConfig } from '@/lib'

export const Roles = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  const { data: config, isLoading } = useQuery({
    queryKey: ['install-app-permissions-config', org?.id, install?.id],
    queryFn: () =>
      getInstallAppPermissionsConfig({
        installId: install.id,
        orgId: org.id,
      }),
    enabled: !!org?.id && !!install?.id,
  })

  const hasRoles =
    config?.provision_role ||
    config?.deprovision_role ||
    config?.maintenance_role ||
    config?.break_glass_roles?.length ||
    config?.custom_roles?.length

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
      ) : hasRoles ? (
        <InstallIAMRoles permissionsConfig={config} />
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
