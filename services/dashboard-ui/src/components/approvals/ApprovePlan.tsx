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
import { APPROVAL_MODAL_COPY } from '@/utils/approval-utils'

interface IApprovePlan {
  step: TWorkflowStep
}

export const ApprovePlanModal = ({ step, ...props }: IApprovePlan & IModal) => {
  const path = usePathname()
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const removePanelByKey = useRemovePanelByKey()
  const { data, error, isLoading, execute } = useServerAction({
    action: approveWorkflowStep,
  })

  const modalCopy = APPROVAL_MODAL_COPY[step.approval.type]

  useServerActionToast({
    data,
    error,
    errorContent: (
      <>
        <Text>There was an error while trying approve these changes.</Text>
        <Text>{error?.error || 'Unknow error occurred.'}</Text>
      </>
    ),
    errorHeading: `Failed to approve changes`,
    onSuccess: () => {
      removePanelByKey(step.id)
      removeModal(props.modalId)
    },
    successContent: (
      <Text>The plan has been approved and the changes are being applied.</Text>
    ),
    successHeading: `Plan approved`,
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
            <Icon variant="Loading" /> Approving plan
          </span>
        ) : (
          'Approve plan'
        ),
        onClick: () => {
          execute({
            body: { note: 'Approved plan', response_type: 'approve' },
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

export const ApprovePlanButton = ({
  step,
  ...props
}: IApprovePlan & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <ApprovePlanModal step={step} />

  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      Approve plan
    </Button>
  )
}
