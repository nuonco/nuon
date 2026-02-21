import { useOrg } from '@/hooks/use-org'
import { useApp } from '@/hooks/use-app'
import { Text } from '@/components/common/Text'
import { PageLayout } from '@/components/layout/PageLayout'
import { PageContent } from '@/components/layout/PageContent'
import { PageHeader } from '@/components/layout/PageHeader'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'

export default function AppRoles() {
  const { org } = useOrg()
  const { app } = useApp()

  return (
    <PageLayout isScrollable>
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name || '' },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name || '' },
          { path: `/${org?.id}/apps/${app?.id}/roles`, text: 'Roles' },
        ]}
      />
      <PageHeader>
        <HeadingGroup>
          <Text variant="h3" weight="stronger" level={1}>
            Roles
          </Text>
        </HeadingGroup>
      </PageHeader>
      <PageContent>
        <Text theme="neutral">Roles content coming soon.</Text>
      </PageContent>
    </PageLayout>
  )
}
