import { Card } from '@/components/common/Card'
import { HealthTimeline } from '@/components/install-health/HealthTimeline'
import { InstallResourcesTable } from '@/components/install-resources/InstallResourcesTable'
import { PageSection } from '@/components/layout/PageSection'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { CustomerManagedSnapshotResources } from '@/components/customer-managed-support/SnapshotResources'
import { isCustomerManagedInstall } from '@/utils/install-utils'

export const Resources = () => {
  const { install } = useInstall()

  return isCustomerManagedInstall(install) ? (
    <CustomerManagedSnapshotResources />
  ) : (
    <ConnectedResources />
  )
}

const ConnectedResources = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  return (
    <PageSection>
      <PageTitle segments={['Resources', install?.name]} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          {
            path: `/${org?.id}/installs/${install?.id}/resources`,
            text: 'Resources',
          },
        ]}
      />
      <SectionHeader
        title="Resources"
        description="Live resources managed by this install's components, with per-resource health."
      />

      <Card>
        <HealthTimeline shouldPoll />
      </Card>

      <InstallResourcesTable shouldPoll />
    </PageSection>
  )
}
