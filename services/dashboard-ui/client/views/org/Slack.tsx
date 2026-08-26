import { useEffect, useRef } from 'react'
import { useSearchParams } from 'react-router'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { ListPage } from '@/components/layout/ListPage'
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

  const toastShown = useRef(false)
  useEffect(() => {
    if (toastShown.current) return
    if (search.get('slack') !== 'installed') return
    toastShown.current = true
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
    <>
      <PageTitle title="Slack" />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org.id}`, text: org?.name },
          { path: `/${org.id}/settings`, text: 'Settings' },
          { path: `/${org.id}/settings/slack`, text: 'Slack' },
        ]}
      />
      <ListPage
        title="Slack"
        description="Receive workflow, workflow step, and approval lifecycle events for this org in Slack channels."
        createAction={<InstallSlackButton />}
      >
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
            <CreateChannelSubscriptionButton variant="secondary" />
          </div>
          <ChannelSubscriptionsTable shouldPoll />
        </div>

        <div className="flex flex-col gap-3">
          <Text variant="base" weight="strong">
            Slash commands
          </Text>
          <Text variant="body" theme="neutral">
            From any channel in a connected workspace, run{' '}
            <span className="font-mono">/nuon subscribe</span> to subscribe the
            current channel,{' '}
            <span className="font-mono">/nuon unsubscribe</span> to remove it,
            or <span className="font-mono">/nuon help</span> for usage.
            Multi-org workspaces require disambiguating the org with{' '}
            <span className="font-mono">
              /nuon subscribe org=&lt;org-id&gt;
            </span>
            .{' '}
            <Link href="https://docs.nuon.co/integrations/slack" isExternal variant="inline">
              Read the docs
            </Link>
          </Text>
        </div>
      </ListPage>
    </>
  )
}
