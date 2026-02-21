import { useParams } from 'react-router-dom'
import { useOrg } from '@/hooks/use-org'
import { useApp } from '@/hooks/use-app'
import { Text } from '@/components/common/Text'
import { PageLayout } from '@/components/layout/PageLayout'
import { PageContent } from '@/components/layout/PageContent'
import { PageHeader } from '@/components/layout/PageHeader'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'

export default function AppActionDetail() {
  const { org } = useOrg()
  const { app } = useApp()
  const { actionId } = useParams()

  return (
    <PageLayout isScrollable>
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name || '' },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name || '' },
          { path: `/${org?.id}/apps/${app?.id}/actions`, text: 'Actions' },
          { path: `/${org?.id}/apps/${app?.id}/actions/${actionId}`, text: 'Action Detail' },
        ]}
      />
      <PageHeader>
        <HeadingGroup>
          <Text variant="h3" weight="stronger" level={1}>
            Action Detail
          </Text>
        </HeadingGroup>
      </PageHeader>
      <PageContent>
        <Text theme="neutral">Action detail content coming soon.</Text>
      </PageContent>
    </PageLayout>
  )
}
