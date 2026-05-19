import { useNavigate } from 'react-router'
import { useMutation } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Toast } from '@/components/surfaces/Toast'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { runRunbook } from '@/lib'
import type { TInstallRunbook } from '@/lib/ctl-api/installs/runbooks'

export const RunRunbookButton = ({
  installRunbook,
  children = 'Run runbook',
  ...props
}: {
  installRunbook: TInstallRunbook
} & IButtonAsButton) => {
  const navigate = useNavigate()
  const { org } = useOrg()
  const { install } = useInstall()
  const { addToast } = useToast()

  const runbookName = installRunbook.runbook?.name ?? 'runbook'
  const runbookId = installRunbook.runbook_id ?? installRunbook.id

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
      const workflowId = result?.install_workflow_id
      if (workflowId) {
        navigate(`/${org!.id}/installs/${install!.id}/workflows/${workflowId}`)
      } else {
        navigate(`/${org!.id}/installs/${install!.id}/runbooks/${runbookId}`)
      }
    },
    onError: () => {
      addToast(
        <Toast
          heading={
            <span className="inline-flex items-center gap-1.5">
              Failed to run <Badge variant="code" size="md">{runbookName}</Badge>
            </span>
          }
          theme="error"
        />
      )
    },
  })

  return (
    <Button
      onClick={() => mutate()}
      disabled={isPending}
      {...props}
    >
      {isPending ? 'Starting...' : children}
    </Button>
  )
}
