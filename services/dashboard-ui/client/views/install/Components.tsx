import { BackToTop } from '@/components/common/BackToTop'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { InstallComponentsTable } from '@/components/install-components/InstallComponentsTable'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'

const CONTAINER_ID = 'install-components-page'

export const Components = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  return (
    <PageSection id={CONTAINER_ID} isScrollable>
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
      <HeadingGroup>
        <Text variant="base" weight="strong">
          Install components
        </Text>
        <Text theme="neutral">
          View and manage all components for this install.
        </Text>
      </HeadingGroup>

      <InstallComponentsTable shouldPoll />
      <BackToTop containerId={CONTAINER_ID} />
    </PageSection>
  )
}
