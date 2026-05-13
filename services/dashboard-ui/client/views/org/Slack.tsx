import { useEffect } from 'react'
import { useSearchParams } from 'react-router'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { PageLayout } from '@/components/layout/PageLayout'
import { PageContent } from '@/components/layout/PageContent'
import { PageHeader } from '@/components/layout/PageHeader'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import {
  ChannelSubscriptionsTable,
  CreateChannelSubscriptionButton,
  InstallationsTable,
  InstallSlackButton,
} from '@/components/slack'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'

export const Slack = () => {
  const { org } = useOrg()
  const [search, setSearch] = useSearchParams()
  const { addToast } = useToast()

  useEffect(() => {
    if (search.get('slack') !== 'installed') return
    addToast(
      <Toast heading="Slack workspace connected" theme="success">
        <Text>
          Subscribe a channel below to start receiving lifecycle events.
        </Text>
      </Toast>
    )
    const next = new URLSearchParams(search)
    next.delete('slack')
    setSearch(next, { replace: true })
  }, [search, setSearch, addToast])

  return (
    <PageLayout className="pb-6">
      <PageTitle title={`Slack | ${org?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org.id}`, text: org?.name },
          { path: `/${org.id}/slack`, text: 'Slack' },
        ]}
      />
      <PageHeader className="flex items-center justify-between">
        <HeadingGroup>
          <Text variant="h3" weight="stronger" level={1}>
            Slack
          </Text>
          <Text theme="neutral">
            Receive workflow, workflow step, and approval lifecycle events for
            this org in Slack channels.
          </Text>
        </HeadingGroup>
        <InstallSlackButton />
      </PageHeader>
      <PageContent>
        <PageSection>
          <div className="flex flex-col gap-4">
            <Text variant="base" weight="strong">
              Connected workspaces
            </Text>
            <InstallationsTable shouldPoll />
          </div>

          <div className="flex flex-col gap-4">
            <div className="flex items-center justify-between">
              <Text variant="base" weight="strong">
                Channel subscriptions
              </Text>
              <CreateChannelSubscriptionButton size="sm" />
            </div>
            <ChannelSubscriptionsTable shouldPoll />
          </div>

          <div className="flex flex-col gap-3">
            <Text variant="base" weight="strong">
              Slash commands
            </Text>
            <Text variant="body" theme="neutral">
              From any channel in a connected workspace, run{' '}
              <span className="font-mono">/nuon subscribe</span> to subscribe
              the current channel,{' '}
              <span className="font-mono">/nuon unsubscribe</span> to remove
              it, or <span className="font-mono">/nuon help</span> for usage.
              Multi-org workspaces require disambiguating the org with{' '}
              <span className="font-mono">/nuon subscribe org=&lt;org-id&gt;</span>
              .{' '}
              <a
                href="https://docs.nuon.co/integrations/slack"
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
