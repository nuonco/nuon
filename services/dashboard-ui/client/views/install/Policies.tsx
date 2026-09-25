import { PolicyReportsFilter } from '@/components/policies/PolicyReportsFilter'
import { InstallPolicyReports } from '@/components/policies/InstallPolicyReports'
import { PageSection } from '@/components/layout/PageSection'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'

export const Policies = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  return (
    <PageSection>
      <PageTitle segments={['Policies', install?.name]} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          {
            path: `/${org?.id}/installs/${install?.id}/policies`,
            text: 'Policies',
          },
        ]}
      />
      <SectionHeader
        title="Policy reports"
        description="Latest policy evaluation for each component in this install."
        actions={<PolicyReportsFilter />}
      />

      <InstallPolicyReports />
    </PageSection>
  )
}
