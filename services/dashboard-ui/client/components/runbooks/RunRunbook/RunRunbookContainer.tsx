import { useState } from 'react'
import { useNavigate } from 'react-router'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { RoleSelector } from '@/components/roles/RoleSelector'
import { type IModal } from '@/components/surfaces/Modal'
import { Toast } from '@/components/surfaces/Toast'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { runRunbook } from '@/lib'
import type {
  TInstallRunbook,
  TRunRunbookBody,
} from '@/lib/ctl-api/installs/runbooks'
import { RunRunbookForm } from './RunRunbookForm'

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
  const [selectedRole, setSelectedRole] = useState('')

  const runbookName = installRunbook.runbook?.name ?? 'runbook'
  const runbookId = installRunbook.runbook_id ?? installRunbook.id

  const { mutate, isPending, error } = useMutation({
    mutationFn: (body: TRunRunbookBody) =>
      runRunbook({
        installId: install!.id,
        runbookId,
        orgId: org!.id,
        body,
      }),
    onSuccess: (result) => {
      addToast(
        <Toast heading="Running runbook" theme="info">
          <Text>Running {runbookName} on {install?.name}.</Text>
        </Toast>
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
  })

  return (
    <RunRunbookForm
      installRunbook={installRunbook}
      isPending={isPending}
      error={error}
      onRun={(body) =>
        mutate({ ...body, ...(selectedRole && { role: selectedRole }) })
      }
      roleSelector={
        <RoleSelector
          installId={install?.id ?? ''}
          operationType="trigger"
          value={selectedRole}
          onChange={setSelectedRole}
          name="role"
        />
      }
      {...props}
    />
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
