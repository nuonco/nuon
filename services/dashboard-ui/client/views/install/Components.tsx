import { InstallComponentsTable } from '@/components/install-components/InstallComponentsTable'
import { ManageAllDropdown } from '@/components/install-components/management/ManageAllDropdown'
import { PageHeader } from '@/components/layout/PageHeader'
import { PageHeadingGroup } from '@/components/layout/PageHeadingGroup'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'

export const Components = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  return (
    <PageSection>
      <PageTitle title={`Components | ${install?.name}`} />
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
      <PageHeader flush>
        <PageHeadingGroup
          title="Install components"
          subtitle="View and manage all components for this install."
          titleProps={{ variant: 'base', weight: 'strong' }}
          headingLevel={2}
        />
        <ManageAllDropdown />
      </PageHeader>

      <InstallComponentsTable shouldPoll />
    </PageSection>
  )
}
