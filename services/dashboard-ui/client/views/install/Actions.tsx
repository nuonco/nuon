import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { InstallActionsTable } from '@/components/actions/InstallActionsTable'
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
      <HeadingGroup>
        <Text variant="base" weight="strong">
          Actions
        </Text>
        <Text theme="neutral">
          View and manage all actions for this install.
        </Text>
      </HeadingGroup>

      <InstallActionsTable shouldPoll />
    </PageSection>
  )
}
