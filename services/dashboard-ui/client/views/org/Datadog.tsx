import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import {
  ConnectionsTable,
  CreateConnectionButton,
  CreateEventSubscriptionButton,
  EventSubscriptionsTable,
  ManagedMonitorsTable,
} from '@/components/datadog'
import { PageContent } from '@/components/layout/PageContent'
import { PageHeader } from '@/components/layout/PageHeader'
import { PageLayout } from '@/components/layout/PageLayout'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useOrg } from '@/hooks/use-org'

export const Datadog = () => {
  const { org } = useOrg()

  return (
    <PageLayout className="pb-6">
      <PageTitle title={`Datadog | ${org?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org.id}`, text: org?.name },
          { path: `/${org.id}/datadog`, text: 'Datadog' },
        ]}
      />
      <PageHeader className="flex items-center justify-between">
        <HeadingGroup>
          <Text variant="h3" weight="stronger" level={1}>
            Datadog
          </Text>
          <Text theme="neutral">
            Stream workflow, step, approval, and drift lifecycle events into
            one or more Datadog tenants — your own monitoring tenant or a
            customer's. Configure event subscriptions to scope what each
            connection receives, and use the one-click "Alert in Datadog"
            action on installs and components to create managed monitors.
          </Text>
        </HeadingGroup>
        <CreateConnectionButton />
      </PageHeader>
      <PageContent>
        <PageSection>
          <div className="flex flex-col gap-4">
            <Text variant="base" weight="strong">
              Connected tenants
            </Text>
            <ConnectionsTable shouldPoll />
          </div>

          <div className="flex flex-col gap-4">
            <div className="flex items-center justify-between">
              <Text variant="base" weight="strong">
                Event subscriptions
              </Text>
              <CreateEventSubscriptionButton size="sm" />
            </div>
            <EventSubscriptionsTable shouldPoll />
          </div>

          <div className="flex flex-col gap-4">
            <Text variant="base" weight="strong">
              Managed monitors
            </Text>
            <ManagedMonitorsTable shouldPoll />
          </div>

          <div className="flex flex-col gap-3">
            <Text variant="base" weight="strong">
              Routing model
            </Text>
            <Text variant="body" theme="neutral">
              Each <strong>connection</strong> binds the org to a Datadog
              tenant — typically one for your team plus one per customer
              who wants visibility. Each <strong>event subscription</strong>{' '}
              is a routing rule on a connection: define a scope (an install,
              component, or label selector) and which event types should
              flow. Multiple connections in the same org can subscribe to
              overlapping scopes — the events fan out to all of them.{' '}
              <a
                href="https://docs.nuon.co/integrations/datadog"
                target="_blank"
                rel="noreferrer noopener"
                className="text-primary-600 dark:text-primary-400 hover:underline inline-flex items-center gap-1"
              >
                Read the docs
                <Icon variant="ArrowSquareOutIcon" size={14} />
              </a>
            </Text>
          </div>
        </PageSection>
      </PageContent>
    </PageLayout>
  )
}
