import { useOutletContext } from 'react-router'
import { TriggerOverview as TriggerOverviewComponent } from '@/components/triggers'
import { PageTitle } from '@/components/navigation/PageTitle'
import type { TTrigger } from '@/types'
export const TriggerOverview = () => {
  const { trigger } = useOutletContext<{ trigger: TTrigger }>()
  return (
    <>
      <PageTitle title={trigger?.name || 'Trigger'} />
      <TriggerOverviewComponent trigger={trigger} />
    </>
  )
}
