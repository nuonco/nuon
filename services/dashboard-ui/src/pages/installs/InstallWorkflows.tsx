import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import { Text } from '@/components/common/Text'
import { PageLayout } from '@/components/layout/PageLayout'
import { PageContent } from '@/components/layout/PageContent'
import { PageHeader } from '@/components/layout/PageHeader'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'

export default function InstallWorkflows() {
  const { org } = useOrg()
  const { install } = useInstall()

  return (
    <PageLayout isScrollable>
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name || '' },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name || '' },
          { path: `/${org?.id}/installs/${install?.id}/workflows`, text: 'Workflows' },
        ]}
      />
      <PageHeader>
        <HeadingGroup>
          <Text variant="h3" weight="stronger" level={1}>
            Workflows
          </Text>
        </HeadingGroup>
      </PageHeader>
      <PageContent>
        <Text theme="neutral">Workflows content coming soon.</Text>
      </PageContent>
    </PageLayout>
  )
}
