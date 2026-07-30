import { useOutletContext } from 'react-router'
import { TriggerOverview as TriggerOverviewComponent } from '@/components/triggers'
import type { TTrigger } from '@/types'
export const TriggerOverview = () => {
  const { trigger } = useOutletContext<{ trigger: TTrigger }>()
  return <TriggerOverviewComponent trigger={trigger} />
}
