import { CreateTriggerButton, TriggersTable } from '@/components/triggers'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageContent } from '@/components/layout/PageContent'
import { PageHeader } from '@/components/layout/PageHeader'
import { PageLayout } from '@/components/layout/PageLayout'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useOrg } from '@/hooks/use-org'
export const Triggers = () => {
  const { org } = useOrg()
  return (
    <PageLayout>
      <PageTitle title={`Triggers | ${org?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/triggers`, text: 'Triggers' },
        ]}
      />
      <PageHeader className="flex items-center justify-between gap-4">
        <HeadingGroup>
          <Text variant="h3" weight="stronger" level={1}>
            Triggers
          </Text>
          <Text theme="neutral">
            Configure inbound providers and inspect their trigger activity.
          </Text>
        </HeadingGroup>
        <CreateTriggerButton />
      </PageHeader>
      <PageContent>
        <PageSection>
          <TriggersTable />
        </PageSection>
      </PageContent>
    </PageLayout>
  )
}
