'use client'

import { usePathname } from 'next/navigation'
import { approveWorkflowStep } from '@/actions/workflows/approve-workflow-step'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useRemovePanelByKey } from '@/hooks/use-remove-panel-by-key'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useServerAction } from '@/hooks/use-server-action'
import { useServerActionToast } from '@/hooks/use-server-action-toast'
import type { TWorkflowStep } from '@/types'
import { RETRY_MODAL_COPY } from '@/utils/approval-utils'

interface IRetryPlan {
  step: TWorkflowStep
}

export const RetryPlanModal = ({ step, ...props }: IRetryPlan & IModal) => {
  const path = usePathname()
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const removePanelByKey = useRemovePanelByKey()
  const { data, error, isLoading, execute } = useServerAction({
    action: approveWorkflowStep,
  })

  const modalCopy = RETRY_MODAL_COPY[step.approval.type]

  useServerActionToast({
    data,
    error,
    errorContent: (
      <>
        <Text>There was an error while retrying these changes.</Text>
        <Text>{error?.error || 'Unknow error occurred.'}</Text>
      </>
    ),
    errorHeading: `Failed to retry changes`,
    onSuccess: () => {
      removePanelByKey(step.id)
      removeModal(props.modalId)
    },
    successContent: (
      <Text>
        A new plan is being generated. Please review the updated changes when
        ready.
      </Text>
    ),
    successHeading: 'Plan retry initiated',
  })

  return (
    <Modal
      heading={
        <Text
          className="inline-flex gap-4 items-center"
          variant="h3"
          weight="stronger"
        >
          {modalCopy.title}
        </Text>
      }
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Retrying plan
          </span>
        ) : (
          'Retry plan'
        ),
        onClick: () => {
          execute({
            body: { note: 'Retry plan', response_type: 'retry' },
            orgId: org.id,
            path,
            workflowId: step.install_workflow_id,
            workflowStepId: step.id,
            approvalId: step?.approval?.id,
          })
        },

        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-1">
        {error ? (
          <Banner theme="error">
            {error?.error ||
              'An error happned, please refresh the page and try again.'}
          </Banner>
        ) : null}
        <Text variant="base" weight="stronger">
          {modalCopy.heading}
        </Text>
        <Text variant="base">{modalCopy.message}</Text>
      </div>
    </Modal>
  )
}

export const RetryPlanButton = ({
  step,
  ...props
}: IRetryPlan & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <RetryPlanModal step={step} />

  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      Retry plan
    </Button>
  )
}
