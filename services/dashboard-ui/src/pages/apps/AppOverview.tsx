import { useOrg } from '@/hooks/use-org'
import { useApp } from '@/hooks/use-app'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageLayout } from '@/components/layout/PageLayout'
import { PageContent } from '@/components/layout/PageContent'
import { PageHeader } from '@/components/layout/PageHeader'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'

export default function AppOverview() {
  const { org } = useOrg()
  const { app } = useApp()

  return (
    <PageLayout isScrollable>
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name || '' },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name || '' },
        ]}
      />
      <PageHeader>
        <HeadingGroup>
          <Text variant="h3" weight="stronger" level={1}>
            {app?.name || 'App'}
          </Text>
        </HeadingGroup>
      </PageHeader>
      <PageContent>
        <Text theme="neutral">App overview coming soon.</Text>
      </PageContent>
    </PageLayout>
  )
}
