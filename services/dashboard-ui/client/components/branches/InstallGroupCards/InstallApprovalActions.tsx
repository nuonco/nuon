import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { ApprovePlanButton } from '@/components/approvals/ApprovePlan'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { approveAllWorkflowSteps, getWorkflowSteps } from '@/lib'
import type { TAPIError, TWorkflowStep } from '@/types'
import type { TTrackInstall } from '@/components/branches/BranchOverview/RolloutTrack'
import { installStatusBucket } from './install-status'

const pendingApprovalSteps = (steps: TWorkflowStep[] | undefined) =>
  (steps ?? []).filter(
    (step) =>
      step.execution_type === 'approval' &&
      step.status?.status === 'approval-awaiting' &&
      !step.approval?.response &&
      step.approval?.id &&
      step.approval?.type !== 'approve-all'
  )

const awaitingInstalls = (installs: TTrackInstall[]) =>
  installs.filter(
    (install) => installStatusBucket(install.status) === 'approval-awaiting'
  )

export const InstallApprovalButton = ({
  install,
  label,
}: {
  install: TTrackInstall
  label?: string
}) => {
  const { org } = useOrg()
  const { data: steps } = useQuery({
    queryKey: ['workflow-steps', org.id, install.workflowId],
    queryFn: () =>
      getWorkflowSteps({
        orgId: org.id,
        workflowId: install.workflowId!,
      }),
    enabled: !!install.workflowId,
  })
  const step = pendingApprovalSteps(steps)[0]
  if (!step) return null
  return (
    <ApprovePlanButton step={step} size="sm" variant="primary">
      {label}
    </ApprovePlanButton>
  )
}

const groupWorkflowIds = (installs: TTrackInstall[]) => [
  ...new Set(
    installs
      .map((install) => install.workflowId)
      .filter((workflowId): workflowId is string => !!workflowId)
  ),
]

const ApproveInstallsModal = ({
  groupName,
  workflowIds,
  onApproved,
  ...props
}: {
  groupName: string
  workflowIds: string[]
  onApproved?: () => void
} & IModal) => {
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
      Promise.all(
        workflowIds.map((workflowId) =>
          approveAllWorkflowSteps({
            orgId: org.id,
            workflowId,
            body: { approval_option: 'approve-all' },
          })
        )
      ),
    onSuccess: () => {
      addToast(
        <Toast heading="All plans approved" theme="success">
          <Text>
            Every install in {groupName} is set to approve all. Those deploys
            are continuing.
          </Text>
        </Toast>
      )
      queryClient.invalidateQueries({ queryKey: ['workflow-steps'] })
      queryClient.invalidateQueries({ queryKey: ['workflow-approvals'] })
      queryClient.invalidateQueries({ queryKey: ['branch-run'] })
      queryClient.invalidateQueries({ queryKey: ['branch-latest-run'] })
      onApproved?.()
      removeModal(props.modalId)
    },
    onError: (err: TAPIError) => {
      addToast(
        <Toast heading="Approval failed" theme="error">
          <Text>{err?.error || 'Unable to approve these plans.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="stronger">
          Approve all plans
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Approving all
          </span>
        ) : (
          'Approve all'
        ),
        onClick: () => execute(),
        disabled: isPending || workflowIds.length === 0,
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-1">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Something went wrong. Try refreshing the page.'}
          </Banner>
        ) : null}
        <Text variant="base">
          Every install workflow in {groupName} will approve every remaining
          plan and continue deploying.
        </Text>
      </div>
    </Modal>
  )
}

export const GroupApproveAllButton = ({
  groupName,
  installs,
  onApproved,
}: {
  groupName: string
  installs: TTrackInstall[]
  onApproved?: () => void
}) => {
  const { addModal } = useSurfaces()
  if (!awaitingInstalls(installs).length) return null
  const workflowIds = groupWorkflowIds(installs)

  return (
    <Button
      size="sm"
      variant="primary"
      onClick={() =>
        addModal(
          <ApproveInstallsModal
            groupName={groupName}
            workflowIds={workflowIds}
            onApproved={onApproved}
          />
        )
      }
    >
      Approve all
    </Button>
  )
}
