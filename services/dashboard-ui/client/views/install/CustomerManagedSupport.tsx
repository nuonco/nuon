import { CustomerManagedSnapshotSupport } from '@/components/customer-managed-support/SnapshotSupport'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'

export const CustomerManagedSupport = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  return (
    <PageSection>
      <PageTitle title={`Support | ${install.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org.id}`, text: org.name },
          { path: `/${org.id}/installs`, text: 'Installs' },
          { path: `/${org.id}/installs/${install.id}`, text: install.name },
          {
            path: `/${org.id}/installs/${install.id}/support`,
            text: 'Support',
          },
        ]}
      />
      <CustomerManagedSnapshotSupport />
    </PageSection>
  )
}
