'use client'

import { usePathname } from 'next/navigation'
import { useEffect } from 'react'
import { cancelWorkflow } from '@/actions/workflows/cancel-workflow'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useServerAction } from '@/hooks/use-server-action'
import { useToast } from '@/hooks/use-toast'
import type { TWorkflow } from '@/types'

interface ICancelWorkflow {
  workflow: TWorkflow
}

export const CancelWorkflowModal = ({
  workflow,
  ...props
}: ICancelWorkflow & IModal) => {
  const path = usePathname()
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const { data, error, isLoading, execute } = useServerAction({
    action: cancelWorkflow,
  })

  useEffect(() => {
    if (data && !error) {
      addToast(
        <Toast theme="info" heading={`${workflow.name} was cancelled.`}>
          <Text>Cancelled the {workflow.type} workflow.</Text>
        </Toast>
      )
      removeModal(props.modalId)
    }

    if (!data && error) {
      addToast(
        <Toast theme="error" heading={`${workflow.name} was not cancelled.`}>
          <Text>
            There was an error while trying to cancel {workflow.type} workflow{' '}
            {workflow.id}.
          </Text>
          <Text>{error?.error || 'Unknow error occurred.'}</Text>
        </Toast>
      )
    }
  }, [data, error])

  return (
    <Modal
      heading={
        <Text
          className="inline-flex gap-4 items-center"
          variant="h3"
          weight="strong"
          theme="error"
        >
          <Icon variant="Warning" size="24" />
          {`Cancel ${workflow?.type} workflow?`}
        </Text>
      }
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Canceling workflow
          </span>
        ) : (
          'Cancel workflow'
        ),
        onClick: () => {
          execute({ orgId: org.id, path, workflowId: workflow.id })
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
        <Text variant="base" weight="strong">
          Are you sure you want to cancel this {workflow.type} workflow?
        </Text>
        <Text variant="base">
          Once a workflow is canceled you can not restart it. You will have to
          trigger a new workflow run.
        </Text>
      </div>
    </Modal>
  )
}

export const CancelWorkflowButton = ({
  workflow,
  ...props
}: ICancelWorkflow & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <CancelWorkflowModal workflow={workflow} />

  return (
    <Button
      variant="danger"
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      Cancel workflow
    </Button>
  )
}
