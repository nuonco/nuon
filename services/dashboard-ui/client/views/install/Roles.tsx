import { PageSection } from '@/components/layout/PageSection'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { InstallRolesTable } from '@/components/roles/InstallRolesTable'
import { CustomerManagedSnapshotRoles } from '@/components/customer-managed-support/SnapshotRoles'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { isCustomerManagedInstall } from '@/utils/install-utils'

const ConnectedRoles = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  return (
    <PageSection>
      <PageTitle segments={['Roles', install?.name]} />
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
      <SectionHeader
        title="Install roles"
        description="View the IAM roles that your install uses to access customer AWS resources."
      />

      <InstallRolesTable />
    </PageSection>
  )
}

export const Roles = () => {
  const { install } = useInstall()
  return isCustomerManagedInstall(install) ? (
    <PageSection>
      <CustomerManagedSnapshotRoles />
    </PageSection>
  ) : (
    <ConnectedRoles />
  )
}
