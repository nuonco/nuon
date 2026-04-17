import { useNavigate } from 'react-router'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useAuth } from '@/hooks/use-auth'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { deprovisionSandbox } from '@/lib'
import { trackEvent } from '@/lib/segment-analytics'
import { DeprovisionSandboxModal } from './DeprovisionSandbox'

export const DeprovisionSandboxModalContainer = ({
  ...props
}: Omit<IModal, 'onSubmit'>) => {
  const navigate = useNavigate()
  const { user } = useAuth()
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const { mutate: execute, isPending, error } = useMutation({
    mutationFn: (params: { body: Parameters<typeof deprovisionSandbox>[0]['body'] }) =>
      deprovisionSandbox({
        body: params.body,
        installId: install.id,
        orgId: org.id,
      }),
    onSuccess: (result) => {
      trackEvent({
        event: 'install_sandbox_deprovision',
        user,
        status: 'ok',
        props: { orgId: org.id, installId: install.id },
      })
      addToast(
        <Toast heading="Deprovision initiated" theme="success">
          <Text>Sandbox deprovision workflow has been started successfully.</Text>
        </Toast>
      )
      queryClient.invalidateQueries({ queryKey: ['workflow-approvals'] })
      queryClient.invalidateQueries({ queryKey: ['active-workflows'] })
      removeModal(props.modalId)
      const workflowId = result.data.workflow_id
      if (workflowId) {
        navigate(`/${org.id}/installs/${install.id}/workflows/${workflowId}`)
      } else {
        navigate(`/${org.id}/installs/${install.id}/workflows`)
      }
    },
    onError: (err: any) => {
      trackEvent({
        event: 'install_sandbox_deprovision',
        user,
        status: 'error',
        props: { orgId: org.id, installId: install.id, err: err?.error },
      })
      addToast(
        <Toast heading="Sandbox deprovision failed" theme="error">
          <Text>Failed to start sandbox deprovision. Please try again.</Text>
        </Toast>
      )
    },
  })

  return (
    <DeprovisionSandboxModal
      installName={install?.name}
      installId={install?.id}
      isPending={isPending}
      error={error}
      onSubmit={({ selectedRole }) => {
        execute({
          body: {
            plan_only: false,
            error_behavior: 'abort',
            ...(selectedRole && { role: selectedRole }),
          },
        })
      }}
      onClose={() => removeModal(props.modalId)}
      {...props}
    />
  )
}

export const DeprovisionSandboxButton = ({
  ...props
}: IButtonAsButton) => {
  const { addModal } = useSurfaces()

  const modal = <DeprovisionSandboxModalContainer />
  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
      className="!text-red-600 dark:!text-red-400"
    >
      {props?.isMenuButton ? null : <Icon variant="BoxArrowDown" />}
      Deprovision sandbox
      {props?.isMenuButton ? <Icon variant="BoxArrowDown" /> : null}
    </Button>
  )
}
