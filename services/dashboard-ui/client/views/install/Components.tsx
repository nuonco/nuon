import { InstallComponentsTable } from '@/components/install-components/InstallComponentsTable'
import { ManageAllDropdown } from '@/components/install-components/management/ManageAllDropdown'
import { PageSection } from '@/components/layout/PageSection'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'

export const Components = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  return (
    <PageSection>
      <PageTitle segments={['Components', install?.name]} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          {
            path: `/${org?.id}/installs/${install?.id}/components`,
            text: 'Components',
          },
        ]}
      />
      <SectionHeader
        title="Install components"
        description="View and manage all components for this install."
        actions={<ManageAllDropdown />}
      />

      <InstallComponentsTable shouldPoll />
    </PageSection>
  )
}
