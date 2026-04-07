import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'

interface ITemporalLink {
  namespace: string
  eventLoopId: string
  isVisible: boolean
}

export const TemporalLink = ({
  namespace,
  eventLoopId,
  isVisible,
}: ITemporalLink) => {
  return isVisible ? (
    <Link
      className="text-xs"
      href={`/admin/temporal/namespaces/${namespace}/workflows/event-loop-${eventLoopId}`}
      target="_blank"
    >
      View in Temporal <Icon variant="ArrowSquareOutIcon" />
    </Link>
  ) : null
}
