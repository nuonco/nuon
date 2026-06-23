import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useWorkflow } from '@/hooks/use-workflow'
import { cancelWorkflow, getWorkflowQueuePosition } from '@/lib'
import type { TAPIError } from '@/types'
import { CancelWorkflowModalComponent as CancelWorkflowModalPresentation } from '@/components/workflows/CancelWorkflow'
import { WorkflowStatusSection } from './WorkflowStatusSection'

export const WorkflowStatusSectionContainer = () => {
  const { workflow } = useWorkflow()
  const { org } = useOrg()
  const { install } = useInstall()
  const { addModal } = useSurfaces()
  const queryClient = useQueryClient()

  const isPending =
    workflow?.status?.status === 'pending' ||
    workflow?.status?.status === 'queued'

  const { data: queuePosition } = useQuery({
    queryKey: ['workflow-queue-position', org?.id, workflow?.id],
    queryFn: () =>
      getWorkflowQueuePosition({
        workflowId: workflow!.id,
        orgId: org!.id,
      }),
    enabled: !!org?.id && !!workflow?.id && isPending,
    refetchInterval: isPending ? 5000 : false,
  })

  const handleCancelWorkflow = (workflowId: string) => {
    const item = queuePosition?.signals_ahead?.find(
      (s) => s.workflow_id === workflowId
    )
    const modal = (
      <InlineCancelWorkflowModal
        workflowId={workflowId}
        workflowType={item?.workflow_type ?? 'unknown'}
        orgId={org?.id ?? ''}
        onSuccess={() => {
          queryClient.invalidateQueries({
            queryKey: ['workflow-queue-position'],
          })
          queryClient.invalidateQueries({ queryKey: ['active-workflows'] })
        }}
      />
    )
    addModal(modal)
  }

  if (!workflow) return null

  return (
    <WorkflowStatusSection
      workflow={workflow}
      queuePosition={isPending ? queuePosition : undefined}
      installId={install?.id}
      orgId={org?.id}
      onCancelWorkflow={handleCancelWorkflow}
    />
  )
}

const InlineCancelWorkflowModal = ({
  workflowId,
  workflowType,
  orgId,
  onSuccess,
  ...props
}: {
  workflowId: string
  workflowType: string
  orgId: string
  onSuccess: () => void
} & Record<string, any>) => {
  const { removeModal } = useSurfaces()

  const {
    mutate: execute,
    isPending,
    error,
  } = useMutation<unknown, TAPIError>({
    mutationFn: () => cancelWorkflow({ orgId, workflowId }),
    onSuccess: () => {
      onSuccess()
      removeModal(props.modalId)
    },
  })

  return (
    <CancelWorkflowModalPresentation
      workflowType={workflowType}
      isPending={isPending}
      error={error}
      onSubmit={() => execute()}
      {...props}
    />
  )
}
