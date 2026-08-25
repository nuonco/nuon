import { useOutletContext, useParams } from 'react-router'
import { TriggerEvents as TriggerEventsComponent } from '@/components/triggers'
import { PageTitle } from '@/components/navigation/PageTitle'
import type { TTrigger } from '@/types'
export const TriggerEvents = () => {
  const { triggerId } = useParams()
  const { trigger } = useOutletContext<{ trigger: TTrigger }>()
  return (
    <>
      <PageTitle title={`${trigger?.name ?? 'Trigger'} events`} />
      {triggerId ? <TriggerEventsComponent triggerId={triggerId} /> : null}
    </>
  )
}
