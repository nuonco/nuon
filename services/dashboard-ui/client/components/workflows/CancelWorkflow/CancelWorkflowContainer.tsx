import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { useWorkflowActions } from '@/hooks/use-workflow-actions'
import { cancelWorkflow } from '@/lib'
import type { TAPIError, TWorkflow } from '@/types'
import { CancelWorkflowModal } from './CancelWorkflow'

interface ICancelWorkflow {
  workflow: TWorkflow
}

export const CancelWorkflowModalContainer = ({
  workflow,
  ...props
}: ICancelWorkflow & Omit<IModal, 'onSubmit'>) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const {
    mutate: execute,
    isPending,
    error,
  } = useMutation<unknown, TAPIError>({
    mutationFn: () =>
      cancelWorkflow({ orgId: org.id, workflowId: workflow.id }),
    onSuccess: () => {
      addToast(
        <Toast heading={`${workflow.name} was cancelled.`} theme="success">
          <Text>Cancelled the {workflow.type} workflow.</Text>
        </Toast>
      )
      queryClient.invalidateQueries({ queryKey: ['workflow-approvals'] })
      queryClient.invalidateQueries({ queryKey: ['active-workflows'] })
      removeModal(props.modalId)
    },
    onError: (err) => {
      addToast(
        <Toast heading={`${workflow.name} was not cancelled.`} theme="error">
          <Text>
            There was an error while trying to cancel {workflow.type} workflow{' '}
            {workflow.id}.
          </Text>
          <Text>{err?.error || 'Unknown error occurred.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <CancelWorkflowModal
      workflowType={workflow.type}
      isPending={isPending}
      error={error}
      onSubmit={() => execute()}
      {...props}
    />
  )
}

export const CancelWorkflowButton = ({
  workflow,
  children,
  ...props
}: ICancelWorkflow & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <CancelWorkflowModalContainer workflow={workflow} />
  const { canShowCancel } = useWorkflowActions(workflow, false)

  return canShowCancel ? (
    <Button
      variant="danger"
      onClick={() => addModal(modal)}
      {...props}
    >
      {children ?? 'Cancel workflow'}
      {props?.isMenuButton ? <Icon variant="StopCircleIcon" /> : null}
    </Button>
  ) : null
}
