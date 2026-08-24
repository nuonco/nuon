import { useNavigate } from 'react-router'
import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { RoleSelector } from '@/components/roles/RoleSelector'
import { useAuth } from '@/hooks/use-auth'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { runAction } from '@/lib'
import { trackEvent } from '@/lib/posthog-analytics'
import type { TAction, TAPIError } from '@/types'
import { InstallActionManualRunModal } from './InstallActionManualRun'

interface IInstallActionManualRunModalContainer extends Omit<IModal, 'heading'> {
  action: TAction
  actionConfigId: string
  isRerun?: boolean
  runEnvVars?: Record<string, string>
}

export const InstallActionManualRunModalContainer = ({
  action,
  actionConfigId,
  isRerun,
  runEnvVars,
  onSubmit: _onSubmit,
  ...props
}: IInstallActionManualRunModalContainer) => {
  const navigate = useNavigate()
  const { user } = useAuth()
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const [selectedRole, setSelectedRole] = useState<string>('')

  const { isPending: isLoading, error, mutate } = useMutation({
    mutationFn: (body: Parameters<typeof runAction>[0]['body'] & { role?: string }) =>
      runAction({ body, installId: install.id, orgId: org.id }),
    onSuccess: (result) => {
      trackEvent({
        event: 'action_run',
        user,
        status: 'ok',
        props: {
          orgId: org?.id,
          installId: install?.id,
          actionConfigId: actionConfigId,
        },
      })
      addToast(
        <Toast heading="Action workflow started" theme="success">
          <Text>The action workflow {action?.name} has been started.</Text>
        </Toast>
      )
      removeModal(props.modalId)
      const workflowId = result.data.workflow_id
      if (workflowId) {
        navigate(`/${org.id}/installs/${install.id}/workflows/${workflowId}`)
      } else {
        navigate(`/${org.id}/installs/${install.id}/workflows`)
      }
    },
    onError: (err: TAPIError) => {
      trackEvent({
        event: 'action_run',
        user,
        status: 'error',
        props: {
          orgId: org?.id,
          installId: install?.id,
          actionConfigId: actionConfigId,
          err: err?.error,
        },
      })
    },
  })

  const handleSubmit = (vars: Record<string, string>) => {
    mutate({
      action_workflow_config_id: actionConfigId,
      ...(vars && Object.keys(vars)?.length > 0 && { run_env_vars: vars }),
      ...(selectedRole && { role: selectedRole }),
    })
  }

  return (
    <InstallActionManualRunModal
      action={action}
      isRerun={isRerun}
      runEnvVars={runEnvVars}
      isLoading={isLoading}
      error={error}
      onSubmit={handleSubmit}
      roleSelector={
        <RoleSelector
          installId={install?.id}
          operationType="trigger"
          principalType="action"
          principalId={action?.id}
          value={selectedRole}
          onChange={setSelectedRole}
          name="role"
        />
      }
      {...props}
    />
  )
}

export const InstallActionManualRunButton = ({
  action,
  actionConfigId,
  isRerun,
  runEnvVars,
  children = 'Run action',
  ...props
}: {
  action: TAction
  actionConfigId: string
  isRerun?: boolean
  runEnvVars?: Record<string, string>
} & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = (
    <InstallActionManualRunModalContainer
      action={action}
      actionConfigId={actionConfigId}
      isRerun={isRerun}
      runEnvVars={runEnvVars}
    />
  )

  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      {children}
    </Button>
  )
}
