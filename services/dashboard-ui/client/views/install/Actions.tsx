import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { InstallActionsTable } from '@/components/actions/InstallActionsTable'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'

export const Actions = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  return (
    <PageSection isScrollable>
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
