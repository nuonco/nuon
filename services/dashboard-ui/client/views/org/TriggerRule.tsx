import { useOutletContext, useParams } from 'react-router'
import { BackLink } from '@/components/common/BackLink'
import { PageTitle } from '@/components/navigation/PageTitle'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { TriggerRuleDetails } from '@/components/triggers'
import type { TTrigger } from '@/types'
export const TriggerRule = () => {
  const { triggerId, ruleId } = useParams()
  const { trigger } = useOutletContext<{ trigger: TTrigger }>()
  return (
    <div className="flex flex-col gap-6">
      <PageTitle title={`${trigger?.name ?? 'Trigger'} rule`} />
      <BackLink />
      <SectionHeader title="Rule details" />
      {triggerId && ruleId ? (
        <TriggerRuleDetails triggerId={triggerId} ruleId={ruleId} />
      ) : null}
    </div>
  )
}
