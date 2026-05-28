import { useNavigate } from 'react-router'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { Toast } from '@/components/surfaces/Toast'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { runRunbook } from '@/lib'
import type { TInstallRunbook } from '@/lib/ctl-api/installs/runbooks'

interface IRunRunbookModal extends IModal {
  installRunbook: TInstallRunbook
}

export const RunRunbookModal = ({
  installRunbook,
  ...props
}: IRunRunbookModal) => {
  const navigate = useNavigate()
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const runbookName = installRunbook.runbook?.name ?? 'runbook'
  const runbookId = installRunbook.runbook_id ?? installRunbook.id
  const steps = installRunbook.runbook?.configs?.[0]?.steps ?? []

  const { mutate, isPending } = useMutation({
    mutationFn: () =>
      runRunbook({
        installId: install!.id,
        runbookId,
        orgId: org!.id,
      }),
    onSuccess: (result) => {
      addToast(
        <Toast
          heading={
            <span className="inline-flex items-center gap-1.5">
              <Badge variant="code" size="md">{runbookName}</Badge> run started
            </span>
          }
          theme="info"
        />
      )
      removeModal(props.modalId)
      queryClient.invalidateQueries({ queryKey: ['install-runbook'] })
      const workflowId = result?.install_workflow_id
      if (workflowId) {
        navigate(`/${org!.id}/installs/${install!.id}/workflows/${workflowId}`)
      } else {
        navigate(`/${org!.id}/installs/${install!.id}/runbooks/${runbookId}`)
      }
    },
    onError: (err: any) => {
      addToast(
        <Toast
          heading={
            <span className="inline-flex items-center gap-1.5">
              Failed to run <Badge variant="code" size="md">{runbookName}</Badge>
            </span>
          }
          theme="error"
        >
          <Text variant="subtext">{err?.error || 'Unknown error occurred'}</Text>
        </Toast>
      )
    },
  })

  return (
    <Modal
      heading={`Run ${runbookName}?`}
      primaryActionTrigger={{
        children: isPending ? (
          <>
            <Icon variant="Loading" className="animate-spin" />
            Running...
          </>
        ) : (
          <>
            Run runbook
            <Icon variant="PlayIcon" />
          </>
        ),
        disabled: isPending,
        onClick: () => mutate(),
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-4">
        <Text>
          This will execute {steps.length} step{steps.length !== 1 ? 's' : ''} in order:
        </Text>
        <ol className="flex flex-col gap-1">
          {steps
            .slice()
            .sort((a, b) => (a.idx ?? 0) - (b.idx ?? 0))
            .map((step, i) => (
              <li key={step.id ?? i} className="flex items-center gap-2">
                <Text as="span" variant="body">
                  {i + 1}. {step.name}
                </Text>
                <Badge variant="code" size="sm" theme="neutral">
                  {step.type}
                </Badge>
              </li>
            ))}
        </ol>
      </div>
    </Modal>
  )
}

export const RunRunbookButton = ({
  installRunbook,
  children = 'Run runbook',
  ...props
}: {
  installRunbook: TInstallRunbook
} & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <RunRunbookModal installRunbook={installRunbook} />

  return (
    <Button onClick={() => addModal(modal)} {...props}>
      {children} <Icon variant="PlayIcon" />
    </Button>
  )
}
