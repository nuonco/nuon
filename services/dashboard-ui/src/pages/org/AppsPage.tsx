import { useOrg } from '@/hooks/use-org'
import { AppsTable } from '@/components/apps/AppsTable'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageLayout } from '@/components/layout/PageLayout'
import { PageContent } from '@/components/layout/PageContent'
import { PageHeader } from '@/components/layout/PageHeader'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'

export default function AppsPage() {
  const { org } = useOrg()

  return (
    <PageLayout isScrollable>
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name || '' },
          { path: `/${org?.id}/apps`, text: 'Apps' },
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
          <AppsTable
            apps={[]}
            pagination={{ limit: 20, offset: 0, hasNext: false }}
            shouldPoll
          />
        </PageSection>
      </PageContent>
    </PageLayout>
  )
}
