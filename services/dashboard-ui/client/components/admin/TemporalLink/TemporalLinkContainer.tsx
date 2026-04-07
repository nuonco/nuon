import { useAuth } from '@/hooks/use-auth'
import { TemporalLink } from './TemporalLink'

export const TemporalLinkContainer = ({
  namespace,
  eventLoopId,
}: {
  namespace: string
  eventLoopId: string
}) => {
  const { user, isLoading } = useAuth()
  const isVisible = !isLoading && !!user?.email?.endsWith('@nuon.co')

  return (
    <TemporalLink
      namespace={namespace}
      eventLoopId={eventLoopId}
      isVisible={isVisible}
    />
  )
}
