import type { ReactNode } from 'react'
import { Button, type TButtonSize } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { SplitButton } from '@/components/common/SplitButton'
import { CancelWorkflowButton } from '@/components/workflows/CancelWorkflow'
import { useWorkflowActions } from '@/hooks/use-workflow-actions'
import type { TWorkflow } from '@/types'

const idleWorkflow = {
  finished: true,
  status: { status: 'success' },
} as TWorkflow

export const DeploymentViewDetails = ({
  workflow,
  onViewDetails,
  size = 'md',
  showArrow = false,
  className,
}: {
  workflow?: TWorkflow
  onViewDetails: () => void
  size?: TButtonSize
  showArrow?: boolean
  className?: string
}) => {
  const { canShowCancel } = useWorkflowActions(workflow ?? idleWorkflow, false)
  const label: ReactNode = showArrow ? (
    <>
      View details <Icon variant="ArrowRightIcon" size={18} />
    </>
  ) : (
    'View details'
  )
  if (!workflow || !canShowCancel) {
    return (
      <Button
        className={className}
        size={size}
        variant="secondary"
        onClick={onViewDetails}
      >
        {label}
      </Button>
    )
  }
  return (
    <SplitButton
      className={className}
      size={size}
      variant="secondary"
      buttonProps={{ children: label, onClick: onViewDetails }}
      dropdownProps={{
        alignment: 'right',
        dropdownClassName: '!border-0 !bg-transparent !shadow-none',
        id: `deployment-${workflow.id}-actions`,
        children: <CancelWorkflowButton workflow={workflow} />,
      }}
    />
  )
}
