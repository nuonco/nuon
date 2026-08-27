import { BranchRunChanges } from '@/components/branches/BranchRunChanges'
import { Text } from '@/components/common/Text'

interface IComparisonStep {
  metadata: Record<string, any>
  status?: string
  appBranchId?: string
  appBranchRunId?: string
}

export const ComparisonStep = ({
  metadata,
  status,
  appBranchId,
  appBranchRunId,
}: IComparisonStep) => {
  const skipReason = metadata.skip_reason as string | undefined

  if (skipReason) {
    return (
      <div className="p-3 bg-cool-grey-100 dark:bg-dark-grey-800 rounded-md">
        <Text variant="base">Differences skipped: {skipReason}</Text>
      </div>
    )
  }

  if (status === 'pending' || status === 'in-progress') {
    return (
      <div className="p-3 bg-cool-grey-100 dark:bg-dark-grey-800 rounded-md">
        <Text variant="base" theme="neutral">
          Computing differences…
        </Text>
      </div>
    )
  }

  if (!appBranchId || !appBranchRunId) {
    return (
      <div className="p-3 bg-cool-grey-100 dark:bg-dark-grey-800 rounded-md">
        <Text variant="base" theme="neutral">
          Run comparison unavailable.
        </Text>
      </div>
    )
  }

  return (
    <BranchRunChanges
      branchId={appBranchId}
      appBranchRunId={appBranchRunId}
      showRunComparison
    />
  )
}
