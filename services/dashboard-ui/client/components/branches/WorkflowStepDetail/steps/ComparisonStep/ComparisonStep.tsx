import { BranchRunChanges } from '@/components/branches/BranchRunChanges'
import { Text } from '@/components/common/Text'
import { StepBlock } from '../../shared/StepLayout'
import { StepStatePlaceholder } from '../../shared/StepStatePlaceholder'

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
      <StepBlock>
        <Text variant="subtext" theme="neutral">
          Differences skipped: {skipReason}
        </Text>
      </StepBlock>
    )
  }

  if (status === 'pending' || status === 'in-progress') {
    return (
      <StepBlock>
        <StepStatePlaceholder
          variant={status === 'pending' ? 'pending' : 'loading'}
        >
          {status === 'pending'
            ? 'Waiting to compute differences'
            : 'Computing differences'}
        </StepStatePlaceholder>
      </StepBlock>
    )
  }

  if (!appBranchId || !appBranchRunId) {
    return (
      <StepBlock>
        <Text variant="subtext" theme="neutral">
          Run comparison unavailable.
        </Text>
      </StepBlock>
    )
  }

  return (
    <StepBlock>
      <BranchRunChanges
        branchId={appBranchId}
        appBranchRunId={appBranchRunId}
        showRunComparison
      />
    </StepBlock>
  )
}
