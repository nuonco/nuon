import { useAuth } from '@/hooks/use-auth'
import { TemporalLink } from './TemporalLink'

export const TemporalLinkContainer = ({
  namespace,
  eventLoopId,
  href,
}: {
  namespace: string
  eventLoopId?: string
  href?: string
}) => {
  const { isAdmin, isLoading } = useAuth()
  const isVisible = !isLoading && !!isAdmin
  const resolvedHref =
    href ??
    `/admin/temporal/namespaces/${namespace}/workflows/event-loop-${eventLoopId}`

  return <TemporalLink href={resolvedHref} isVisible={isVisible} />
}
