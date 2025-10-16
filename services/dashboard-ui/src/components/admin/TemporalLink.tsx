'use client'

import { useUser } from '@auth0/nextjs-auth0'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'

export const TemporalLink = ({
  namespace,
  eventLoopId,
}: {
  namespace: string
  eventLoopId: string
}) => {
  const { user, isLoading } = useUser()

  return !isLoading && user?.email?.endsWith('@nuon.co') ? (
    <Link
      className="text-sm"
      href={`/admin/temporal/namespaces/${namespace}/workflows/event-loop-${eventLoopId}`}
      target="_blank"
    >
      View in Temporal <Icon variant="ArrowSquareOutIcon" />
    </Link>
  ) : null
}
