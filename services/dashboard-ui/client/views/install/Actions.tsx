import { InstallActionsTable } from '@/components/actions/InstallActionsTable'
import { RunAdhocActionButton } from '@/components/installs/management/RunAdhocAction'
import { PageHeader } from '@/components/layout/PageHeader'
import { PageHeadingGroup } from '@/components/layout/PageHeadingGroup'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'

export const Actions = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  return (
    <PageSection>
      <PageTitle title={`Actions | ${install?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          {
            path: `/${org?.id}/installs/${install?.id}/actions`,
            text: 'Actions',
          },
        ]}
      />
      <PageHeader flush>
        <PageHeadingGroup
          title="Actions"
          subtitle="View and manage all actions for this install."
          titleProps={{ variant: 'base', weight: 'strong' }}
          headingLevel={2}
        />
        <RunAdhocActionButton variant="primary" />
      </PageHeader>

      <InstallActionsTable shouldPoll />
    </PageSection>
  )
}
