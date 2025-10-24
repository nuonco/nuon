'use client'

import { usePathname } from 'next/navigation'
import { approveAllWorkflowSteps } from '@/actions/workflows/approve-all-workflow-steps'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useServerAction } from '@/hooks/use-server-action'
import { useServerActionToast } from '@/hooks/use-server-action-toast'
import type { TWorkflow } from '@/types'
import { toSentenceCase } from '@/utils/string-utils'

interface IApproveAll {
  workflow: TWorkflow
}

export const ApproveAllModal = ({
  workflow,
  ...props
}: IApproveAll & IModal) => {
  const path = usePathname()
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { data, error, isLoading, execute } = useServerAction({
    action: approveAllWorkflowSteps,
  })

  useServerActionToast({
    data,
    error,
    errorContent: (
      <>
        <Text>
          There was an error while trying approve all changes to {workflow.type}{' '}
          workflow {workflow.id}.
        </Text>
        <Text>{error?.error || 'Unknow error occurred.'}</Text>
      </>
    ),
    errorHeading: 'Failed to approve changes',
    onSuccess: () => {
      removeModal(props.modalId)
    },
    successContent: (
      <Text>
        All proposed changes have been approved and are being applied across the
        workflow.
      </Text>
    ),
    successHeading: 'All plans approved',
  })

  return (
    <Modal
      heading={
        <Text
          className="inline-flex gap-4 items-center"
          variant="h3"
          weight="stronger"
        >
          Approve all plans?
        </Text>
      }
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Approving changes
          </span>
        ) : (
          'Approve all'
        ),
        onClick: () => {
          execute({
            body: { approval_option: 'approve-all' },
            orgId: org.id,
            path,
            workflowId: workflow.id,
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
          Are you sure you want to approve all proposed changes?
        </Text>
        <Text variant="base">
          Approving all plans will immediately apply every set of outlined
          changes across the workflow.
        </Text>
        <Text className="mt-3" variant="base" weight="stronger">
          Step to approve
        </Text>
        <div className="flex flex-wrap gap-2">
          {workflow?.steps
            ?.filter(
              (s) => s?.execution_type === 'approval' && !s?.approval?.response
            )
            .map((s) => (
              <Badge variant="code" key={s?.id} size="sm">
                {toSentenceCase(s?.name)}
              </Badge>
            ))}
        </div>
      </div>
    </Modal>
  )
}

export const ApproveAllButton = ({
  workflow,
  ...props
}: IApproveAll & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <ApproveAllModal workflow={workflow} />

  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      Approve all
    </Button>
  )
}
