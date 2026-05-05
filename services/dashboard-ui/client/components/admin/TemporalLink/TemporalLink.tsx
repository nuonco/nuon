import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { useAuth } from '@/hooks/use-auth'

export const TemporalLink = ({
  namespace,
  eventLoopId,
  href,
}: {
  namespace: string
  eventLoopId?: string
  href?: string
}) => {
  const { isAdmin } = useAuth()

  if (!isAdmin) {
    return null
  }

  const resolvedHref =
    href ??
    `/admin/temporal/namespaces/${namespace}/workflows/event-loop-${eventLoopId}`

  return (
    <Link
      className="text-xs inline-flex items-center gap-1"
      href={resolvedHref}
      target="_blank"
    >
      admin <Icon variant="ArrowSquareOutIcon" size="14" />
    </Link>
  )
}
