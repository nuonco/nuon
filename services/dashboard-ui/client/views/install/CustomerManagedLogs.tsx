import { Navigate } from 'react-router'
import { CustomerManagedSnapshotLogs } from '@/components/customer-managed-support/SnapshotLogs'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { isCustomerManagedInstall } from '@/utils/install-utils'

export const CustomerManagedLogs = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  if (!isCustomerManagedInstall(install)) {
    return <Navigate to={`/${org.id}/installs/${install.id}/runner`} replace />
  }

  return (
    <PageSection>
      <PageTitle title={`Logs | ${install.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org.id}`, text: org.name },
          { path: `/${org.id}/installs`, text: 'Installs' },
          { path: `/${org.id}/installs/${install.id}`, text: install.name },
          { path: `/${org.id}/installs/${install.id}/logs`, text: 'Logs' },
        ]}
      />
      <CustomerManagedSnapshotLogs />
    </PageSection>
  )
}
