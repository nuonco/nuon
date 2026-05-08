import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getSlackChannelSubscriptions, getSlackOrgLinks } from '@/lib'
import { ChannelSubscriptionsTable } from './ChannelSubscriptionsTable'

export const ChannelSubscriptionsTableContainer = ({
  pollInterval = 30000,
  shouldPoll = true,
}: {
  pollInterval?: number
  shouldPoll?: boolean
} = {}) => {
  const { org } = useOrg()

  const subsQuery = useQuery({
    queryKey: ['slack-channel-subscriptions', org.id],
    queryFn: () => getSlackChannelSubscriptions({ orgId: org.id }),
    refetchInterval: shouldPoll ? pollInterval : false,
  })

  const linksQuery = useQuery({
    queryKey: ['slack-org-links', org.id],
    queryFn: () => getSlackOrgLinks({ orgId: org.id }),
  })

  return (
    <ChannelSubscriptionsTable
      data={subsQuery.data ?? []}
      links={linksQuery.data ?? []}
      isLoading={subsQuery.isLoading || linksQuery.isLoading}
    />
  )
}
