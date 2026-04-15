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
  const { user, isLoading } = useAuth()
  const isVisible = !isLoading && !!user?.email?.endsWith('@nuon.co')
  const resolvedHref =
    href ??
    `/admin/temporal/namespaces/${namespace}/workflows/event-loop-${eventLoopId}`

  return <TemporalLink href={resolvedHref} isVisible={isVisible} />
}
