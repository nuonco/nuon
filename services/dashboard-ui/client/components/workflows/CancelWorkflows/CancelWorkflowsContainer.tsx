import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { cancelWorkflows } from '@/lib'
import type { TAPIError } from '@/types'
import { CancelWorkflowsModal } from './CancelWorkflows'

interface ICancelWorkflowsContainer extends Omit<IModal, 'onSubmit'> {
  workflowIds: string[]
  onComplete: () => void
}

export const CancelWorkflowsContainer = ({
  workflowIds,
  onComplete,
  ...props
}: ICancelWorkflowsContainer) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const {
    mutate: execute,
    isPending,
    error,
    data: results,
  } = useMutation<
    { cancelled: string[]; errors?: { workflow_id: string; error: string }[] },
    TAPIError
  >({
    mutationFn: () =>
      cancelWorkflows({ orgId: org.id, workflowIds }),
    onSuccess: (data) => {
      const cancelledCount = data.cancelled?.length ?? 0
      const errorCount = data.errors?.length ?? 0

      if (errorCount === 0) {
        addToast(
          <Toast heading={`${cancelledCount} workflow${cancelledCount === 1 ? '' : 's'} cancelled`} theme="success">
            <Text>All selected workflows were cancelled.</Text>
          </Toast>
        )
        removeModal(props.modalId)
        onComplete()
      } else if (cancelledCount > 0) {
        addToast(
          <Toast heading={`${cancelledCount} cancelled, ${errorCount} failed`} theme="warn">
            <Text>Some workflows could not be cancelled.</Text>
          </Toast>
        )
      } else {
        addToast(
          <Toast heading="Workflow cancellation failed" theme="error">
            <Text>None of the selected workflows could be cancelled.</Text>
          </Toast>
        )
      }

      queryClient.invalidateQueries({ queryKey: ['workflow-approvals'] })
      queryClient.invalidateQueries({ queryKey: ['active-workflows'] })
      queryClient.invalidateQueries({ queryKey: ['install-active-workflows'] })
      queryClient.invalidateQueries({ queryKey: ['install-workflows'] })
    },
    onError: (err) => {
      addToast(
        <Toast heading="Workflow cancellation failed" theme="error">
          <Text>{err?.error || 'An unknown error occurred.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <CancelWorkflowsModal
      count={workflowIds.length}
      isPending={isPending}
      error={error?.error}
      cancelResults={results}
      onSubmit={() => execute()}
      {...props}
    />
  )
}
