import { useQuery } from '@tanstack/react-query'
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
    return <Text theme="error">Rule loading failed.</Text>
  return <TriggerRuleDetails orgId={org?.id ?? ''} rule={query.data} />
}
