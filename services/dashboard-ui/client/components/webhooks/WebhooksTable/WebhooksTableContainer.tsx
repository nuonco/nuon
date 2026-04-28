import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getCurrentOrgWebhooks } from '@/lib'
import { WebhooksTable } from './WebhooksTable'

export const WebhooksTableContainer = ({
  pollInterval = 20000,
  shouldPoll = true,
}: {
  pollInterval?: number
  shouldPoll?: boolean
} = {}) => {
  const { org } = useOrg()

  const { data, isLoading } = useQuery({
    queryKey: ['webhooks', org.id],
    queryFn: () => getCurrentOrgWebhooks({ orgId: org.id }),
    refetchInterval: shouldPoll ? pollInterval : false,
  })

  return <WebhooksTable data={data ?? []} isLoading={isLoading} />
}
