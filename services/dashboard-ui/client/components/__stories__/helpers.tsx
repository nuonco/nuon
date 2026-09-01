import type { ReactNode } from 'react'
import { Button } from '@/components/common/Button'
import { useSurfaces } from '@/hooks/use-surfaces'
import type { IModal } from '@/components/surfaces/Modal'
import type { IPanel } from '@/components/surfaces/Panel'
import { StepCard } from '@/components/branches/WorkflowStepDetail/StepCard'
import type { TInstallWorkflowStep } from '@/types'

export const ModalStory = ({
  children,
  label,
}: {
  children: ReactNode
  label?: string
}) => {
  const { addModal } = useSurfaces()

  return (
    <Button
      variant="primary"
      onClick={() => addModal(children as React.ReactElement<IModal>)}
    >
      {label || 'Open modal'}
    </Button>
  )
}

export const PanelStory = ({
  children,
  label,
}: {
  children: ReactNode
  label?: string
}) => {
  const { addPanel } = useSurfaces()

  return (
    <Button
      variant="primary"
      onClick={() => addPanel(children as React.ReactElement<IPanel>)}
    >
      {label || 'Open panel'}
    </Button>
  )
}

export const StepCardStory = ({
  children,
  name = 'step details',
  status = 'success',
}: {
  children: ReactNode
  name?: string
  status?: string
}) => (
  <StepCard
    step={
      {
        id: 'step-preview',
        name,
        group_idx: 1,
        started_at: '2024-06-15T10:30:00Z',
        execution_time: 25200000000,
        status: { status },
      } as TInstallWorkflowStep
    }
  >
    {children}
  </StepCard>
)
