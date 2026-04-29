import { useOutletContext } from 'react-router'
import { Plan } from '@/components/approvals/Plan'
import { EmptyState } from '@/components/common/EmptyState'
import type { TSandboxRunOutletContext } from './types'

export const SandboxRunPlanTab = () => {
  const { step } = useOutletContext<TSandboxRunOutletContext>()

  if (!step?.approval) {
    return (
      <EmptyState
        variant="table"
        emptyTitle="No plan available"
        emptyMessage="No approval plan has been generated for this sandbox run."
      />
    )
  }

  return <Plan step={step} />
}
