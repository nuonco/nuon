import { useNavigate } from 'react-router'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useAuth } from '@/hooks/use-auth'
import { Badge } from '@/components/common/Badge'
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
        <Toast heading="Sandbox deprovision started" theme="info">
          <Text>Deprovisioning <Badge variant="code" size="md">{install.name}</Badge> sandbox. This may take a few minutes.</Text>
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
          <Text>Unable to deprovision <Badge variant="code" size="md">{install.name}</Badge> sandbox.</Text>
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
      variant="danger"
    >
      {props?.isMenuButton ? null : <Icon variant="BoxArrowDownIcon" />}
      Deprovision sandbox
      {props?.isMenuButton ? <Icon variant="BoxArrowDownIcon" /> : null}
    </Button>
  )
}
