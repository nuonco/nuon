import { useOutletContext, useParams } from 'react-router'
import { EventDetails } from '@/components/triggers'
import { PageTitle } from '@/components/navigation/PageTitle'
import type { TTrigger } from '@/types'

export const TriggerEvent = () => {
  const { triggerId } = useParams()
  const { trigger } = useOutletContext<{ trigger: TTrigger }>()
  return (
    <>
      <PageTitle title={`${trigger?.name ?? 'Trigger'} event`} />
      <EventDetails expectedTriggerId={triggerId} />
    </>
  )
}
