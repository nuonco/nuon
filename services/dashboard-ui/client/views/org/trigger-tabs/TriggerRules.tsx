import { useOutletContext, useParams } from 'react-router'
import { TriggerRules as TriggerRulesComponent } from '@/components/triggers'
import { PageTitle } from '@/components/navigation/PageTitle'
import type { TTrigger } from '@/types'
export const TriggerRules = () => {
  const { triggerId } = useParams()
  const { trigger } = useOutletContext<{ trigger: TTrigger }>()
  return (
    <>
      <PageTitle title={`${trigger?.name ?? 'Trigger'} rules`} />
      {triggerId ? <TriggerRulesComponent triggerId={triggerId} /> : null}
    </>
  )
}
