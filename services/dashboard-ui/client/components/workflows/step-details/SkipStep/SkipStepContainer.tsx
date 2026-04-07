import { useMutation } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useRemovePanelByKey } from '@/hooks/use-remove-panel-by-key'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { retryWorkflowStep } from '@/lib'
import type { TAPIError, TWorkflowStep } from '@/types'
import { toSentenceCase } from '@/utils/string-utils'
import { SkipStepModal } from './SkipStep'

interface ISkipStep {
  step: TWorkflowStep
}

export const SkipStepModalContainer = ({
  step,
  ...props
}: ISkipStep & Omit<IModal, 'onSubmit'>) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const removePanelByKey = useRemovePanelByKey()

  const {
    mutate: execute,
    isPending,
    error,
  } = useMutation<unknown, TAPIError>({
    mutationFn: () =>
      retryWorkflowStep({
        body: { operation: 'skip-step', step_id: step.id },
        orgId: org.id,
        workflowId: step.install_workflow_id,
      }),
    onSuccess: () => {
      addToast(
        <Toast heading="Step skipped" theme="success">
          <Text>
            {toSentenceCase(step.name)} was skipped. The workflow will continue
            with the remaining steps.
          </Text>
        </Toast>
      )
      removePanelByKey(step.id)
      removeModal(props.modalId)
    },
    onError: (err) => {
      addToast(
        <Toast heading="Failed to skip step" theme="error">
          <Text>There was an error while skipping this step.</Text>
          <Text>{err?.error || 'Unknown error occurred.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <SkipStepModal
      isPending={isPending}
      error={error}
      onSubmit={() => execute()}
      {...props}
    />
  )
}

export const SkipStepButton = ({
  step,
  ...props
}: ISkipStep & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <SkipStepModalContainer step={step} />

  return (
    <Button onClick={() => addModal(modal)} {...props}>
      Skip step
    </Button>
  )
}
