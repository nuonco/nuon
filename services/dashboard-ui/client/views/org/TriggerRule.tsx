import { useParams } from 'react-router'
import { BackLink } from '@/components/common/BackLink'
import { Text } from '@/components/common/Text'
import { TriggerRuleDetails } from '@/components/triggers'
export const TriggerRule = () => {
  const { triggerId, ruleId } = useParams()
  return (
    <div className="flex flex-col gap-6">
      <BackLink />
      <Text variant="h3" weight="strong">
        Rule details
      </Text>
      {triggerId && ruleId ? (
        <TriggerRuleDetails triggerId={triggerId} ruleId={ruleId} />
      ) : null}
    </div>
  )
}
