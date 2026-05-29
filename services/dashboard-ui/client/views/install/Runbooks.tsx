import { InstallRunbooksTable } from '@/components/runbooks/InstallRunbooksTable'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'

export const Runbooks = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  return (
    <PageSection>
      <PageTitle title={`Runbooks | ${install?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          {
            path: `/${org?.id}/installs/${install?.id}/runbooks`,
            text: 'Runbooks',
          },
        ]}
      />
      <HeadingGroup>
        <Text variant="base" weight="strong">
          Runbooks
        </Text>
        <Text variant="subtext" theme="neutral">
          View and run operational procedures for this install.
        </Text>
      </HeadingGroup>
      <InstallRunbooksTable shouldPoll />
    </PageSection>
  )
}
