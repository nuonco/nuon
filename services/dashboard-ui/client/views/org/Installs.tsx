import { BackToTop } from '@/components/common/BackToTop'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { InstallsTable } from '@/components/installs/InstallsTable'
import { PageLayout } from '@/components/layout/PageLayout'
import { PageContent } from '@/components/layout/PageContent'
import { PageHeader } from '@/components/layout/PageHeader'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useOrg } from '@/hooks/use-org'

const CONTAINER_ID = 'org-installs-page'

export const Installs = () => {
  const { org } = useOrg()

  return (
    <PageLayout className="pb-6" id={CONTAINER_ID} isScrollable>
      <PageTitle title={`Installs | ${org?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          {
            path: `/${org.id}`,
            text: org?.name,
          },
          {
            path: `/${org.id}/installs`,
            text: 'Installs',
          },
        ]}
      />
      <PageHeader>
        <HeadingGroup>
          <Text variant="h3" weight="stronger" level={1}>
            Installs
          </Text>
          <Text theme="neutral">
            View and manage all deployed installs here.
          </Text>
        </HeadingGroup>
      </PageHeader>
      <PageContent>
        <PageSection>
          <InstallsTable shouldPoll />
        </PageSection>
      </PageContent>
      <BackToTop containerId={CONTAINER_ID} />
    </PageLayout>
  )
}
