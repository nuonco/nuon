import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getTriggerRules } from '@/lib'
import { TriggerRules } from './TriggerRules'
export const TriggerRulesContainer = ({ triggerId }: { triggerId: string }) => {
  const { org } = useOrg()
  const query = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['event-trigger-rules', org?.id, triggerId],
    queryFn: () => getTriggerRules({ triggerId: triggerId, orgId: org!.id }),
    enabled: !!org?.id && !!triggerId,
  })
  return (
    <TriggerRules
      data={query.data ?? []}
      hasError={query.isError}
      isLoading={query.isLoading}
      onRetry={() => void query.refetch()}
      orgId={org?.id ?? ''}
      triggerId={triggerId}
    />
  )
}
