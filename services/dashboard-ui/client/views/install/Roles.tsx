import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { InstallRolesTable } from '@/components/roles/InstallRolesTable'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'

export const Roles = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  return (
    <PageSection>
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

      <InstallRolesTable />
    </PageSection>
  )
}
