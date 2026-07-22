import { useParams } from 'react-router'
import { TriggerRules as TriggerRulesComponent } from '@/components/triggers'
export const TriggerRules = () => {
  const { triggerId } = useParams()
  return triggerId ? <TriggerRulesComponent triggerId={triggerId} /> : null
}
