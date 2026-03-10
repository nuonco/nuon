import { AppsTable } from '@/components/apps/AppsTable'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageContent } from '@/components/layout/PageContent'
import { PageHeader } from '@/components/layout/PageHeader'
import { PageLayout } from '@/components/layout/PageLayout'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useOrg } from '@/hooks/use-org'

export const Apps = () => {
  const { org } = useOrg()

  return (
    <PageLayout isScrollable>
      <PageTitle title={`Apps | ${org?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          {
            path: `/${org.id}`,
            text: org?.name,
          },
          {
            path: `/${org.id}/apps`,
            text: 'Apps',
          },
        ]}
      />

      <PageHeader>
        <HeadingGroup>
          <Text variant="h3" weight="stronger" level={1}>
            Apps
          </Text>
          <Text theme="neutral">Manage your applications here.</Text>
        </HeadingGroup>
      </PageHeader>

      <PageContent>
        <PageSection>
          <AppsTable shouldPoll />
        </PageSection>
      </PageContent>
    </PageLayout>
  )
}
