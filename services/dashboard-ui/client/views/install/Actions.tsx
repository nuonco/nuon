import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { InstallActionsTable } from '@/components/actions/InstallActionsTable'
import { RunAdhocActionButton } from '@/components/installs/management/RunAdhocAction'
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
      <div className="flex items-start justify-between gap-4">
        <HeadingGroup>
          <Text variant="base" weight="strong">
            Actions
          </Text>
          <Text variant="subtext" theme="neutral">
            View and manage all actions for this install.
          </Text>
        </HeadingGroup>
        <div className="shrink-0">
          <RunAdhocActionButton />
        </div>
      </div>

      <InstallActionsTable shouldPoll />
    </PageSection>
  )
}
