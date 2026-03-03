import { useMutation } from '@tanstack/react-query'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useRemovePanelByKey } from '@/hooks/use-remove-panel-by-key'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { retryWorkflowStep } from '@/lib'
import type { TAPIError, TWorkflowStep } from '@/types'
import { toSentenceCase } from '@/utils/string-utils'

interface ISkipStep {
  step: TWorkflowStep
}

export const SkipStepModal = ({ step, ...props }: ISkipStep & IModal) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const removePanelByKey = useRemovePanelByKey()

  const { mutate: execute, isPending: isLoading, error } = useMutation<unknown, TAPIError>({
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
            {toSentenceCase(step.name)} was skipped. The workflow will continue with the remaining
            steps.
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
    <Modal
      heading={
        <Text className="inline-flex gap-4 items-center" variant="h3" weight="stronger">
          Skip step?
        </Text>
      }
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Skipping step
          </span>
        ) : (
          'Skip step'
        ),
        onClick: () => execute(),
        disabled: isLoading,
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-1">
        {error ? (
          <Banner theme="error">
            {error?.error || 'An error happened, please refresh the page and try again.'}
          </Banner>
        ) : null}
        <Text variant="base" weight="stronger">
          Are you sure you want to skip this step?
        </Text>
        <Text variant="base">
          Skipping will bypass this step and continue the workflow with the remaining steps. Any
          actions or changes from this step will not be applied.
        </Text>
      </div>
    </Modal>
  )
}

export const SkipStepButton = ({ step, ...props }: ISkipStep & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <SkipStepModal step={step} />

  return (
    <Button onClick={() => addModal(modal)} {...props}>
      Skip step
    </Button>
  )
}
