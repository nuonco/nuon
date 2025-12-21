'use client'

import { useAuth } from '@/hooks/use-auth'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'

export const TemporalLink = ({
  namespace,
  eventLoopId,
  skipPrefix = false,
}: {
  namespace: string
  eventLoopId: string
  skipPrefix?: boolean
}) => {
  const { user, isLoading } = useAuth()

  const workflowId = skipPrefix ? eventLoopId : `event-loop-${eventLoopId}`

  return !isLoading && user?.email?.endsWith('@nuon.co') ? (
    <Link
      className="text-xs"
      href={`/admin/temporal/namespaces/${namespace}/workflows/${workflowId}`}
      target="_blank"
    >
      View in Temporal <Icon variant="ArrowSquareOutIcon" />
    </Link>
  ) : null
}
