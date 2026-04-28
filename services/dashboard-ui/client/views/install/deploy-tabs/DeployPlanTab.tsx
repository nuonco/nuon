import { useOutletContext } from 'react-router'
import { Plan } from '@/components/approvals/Plan'
import { EmptyState } from '@/components/common/EmptyState'
import type { TDeployOutletContext } from './types'

export const DeployPlanTab = () => {
  const { step } = useOutletContext<TDeployOutletContext>()

  if (!step?.approval) {
    return (
      <EmptyState
        variant="table"
        emptyTitle="No plan available"
        emptyMessage="No approval plan has been generated for this deploy."
      />
    )
  }

  return <Plan step={step} />
}
