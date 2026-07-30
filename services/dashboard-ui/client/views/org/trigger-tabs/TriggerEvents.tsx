import { useParams } from 'react-router'
import { TriggerEvents as TriggerEventsComponent } from '@/components/triggers'
export const TriggerEvents = () => {
  const { triggerId } = useParams()
  return triggerId ? <TriggerEventsComponent triggerId={triggerId} /> : null
}
