import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { approveAllWorkflowSteps } from '@/lib'
import type { TWorkflow } from '@/types'
import { ApproveAllModal } from './ApproveAll'

interface IApproveAll {
  workflow: TWorkflow
}

export const ApproveAllModalContainer = ({
  workflow,
  ...props
}: IApproveAll & Omit<IModal, 'onSubmit'>) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const {
    mutate: execute,
    isPending,
    error,
  } = useMutation({
    mutationFn: () =>
      approveAllWorkflowSteps({
        body: { approval_option: 'approve-all' },
        orgId: org.id,
        workflowId: workflow.id,
      }),
    onSuccess: () => {
      addToast(
        <Toast heading="All plans approved" theme="success">
          <Text>
            All proposed changes have been approved and are being applied across
            the workflow.
          </Text>
        </Toast>
      )
      queryClient.invalidateQueries({ queryKey: ['workflow-approvals'] })
      queryClient.invalidateQueries({ queryKey: ['active-workflows'] })
      queryClient.invalidateQueries({ queryKey: ['workflow-steps'] })
      removeModal(props.modalId)
    },
    onError: (err: any) => {
      addToast(
        <Toast heading="Failed to approve changes" theme="error">
          <Text>
            There was an error while trying approve all changes to{' '}
            {workflow.type} workflow {workflow.id}.
          </Text>
          <Text>{err?.error || 'Unknown error occurred.'}</Text>
        </Toast>
      )
    },
  })

  const pendingSteps = (workflow?.steps ?? [])
    .filter(
      (s) => s?.execution_type === 'approval' && !s?.approval?.response
    )
    .map((s) => ({ id: s.id, name: s.name }))

  return (
    <ApproveAllModal
      pendingSteps={pendingSteps}
      isPending={isPending}
      error={error}
      onSubmit={() => execute()}
      {...props}
    />
  )
}

export const ApproveAllButton = ({
  workflow,
  ...props
}: IApproveAll & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <ApproveAllModalContainer workflow={workflow} />

  return (
    <Button onClick={() => addModal(modal)} {...props}>
      Approve all
    </Button>
  )
}
