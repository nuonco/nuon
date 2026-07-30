import { EventDetails } from '@/components/triggers'
import { useParams } from 'react-router'

export const TriggerEvent = () => {
  const { triggerId } = useParams()
  return <EventDetails expectedTriggerId={triggerId} />
}
