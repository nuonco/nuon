import { CustomerManagedSnapshotBundles } from '@/components/customer-managed-support/SnapshotBundles'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'

export const CustomerManagedBundles = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  return (
    <PageSection>
      <PageTitle title={`Bundles | ${install.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org.id}`, text: org.name },
          { path: `/${org.id}/installs`, text: 'Installs' },
          { path: `/${org.id}/installs/${install.id}`, text: install.name },
          {
            path: `/${org.id}/installs/${install.id}/bundles`,
            text: 'Bundles',
          },
        ]}
      />
      <CustomerManagedSnapshotBundles />
    </PageSection>
  )
}
