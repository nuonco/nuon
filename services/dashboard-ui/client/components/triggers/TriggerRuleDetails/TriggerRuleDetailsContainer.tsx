import { useQuery } from '@tanstack/react-query'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { useOrg } from '@/hooks/use-org'
import { getTriggerRule } from '@/lib'
import { TriggerRuleDetails } from './TriggerRuleDetails'
export const TriggerRuleDetailsContainer = ({
  ruleId,
  triggerId,
}: {
  ruleId: string
  triggerId: string
}) => {
  const { org } = useOrg()
  const query = useQuery({
    queryKey: ['event-trigger-rule', org?.id, triggerId, ruleId],
    queryFn: () =>
      getTriggerRule({ orgId: org!.id, triggerId: triggerId, ruleId }),
    enabled: !!org?.id && !!triggerId && !!ruleId,
  })
  if (query.isLoading) return <Text theme="neutral">Loading rule...</Text>
  if (query.error || !query.data)
    return (
      <div className="flex flex-col items-start gap-3">
        <Text theme="error">Rule loading failed.</Text>
        <Button variant="secondary" onClick={() => void query.refetch()}>
          <Icon variant="ArrowClockwiseIcon" />
          Retry loading rule
        </Button>
      </div>
    )
  return <TriggerRuleDetails orgId={org?.id ?? ''} rule={query.data} />
}
