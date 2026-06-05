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
import { reprovisionSandbox } from '@/lib'
import { trackEvent } from '@/lib/segment-analytics'
import { ReprovisionSandboxModal } from './ReprovisionSandbox'

export const ReprovisionSandboxModalContainer = ({
  onSubmit: _onSubmit,
  ...props
}: IModal) => {
  const navigate = useNavigate()
  const { user } = useAuth()
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const { mutate: execute, isPending, error } = useMutation({
    mutationFn: (params: { body: Parameters<typeof reprovisionSandbox>[0]['body'] }) =>
      reprovisionSandbox({
        body: params.body,
        installId: install.id,
        orgId: org.id,
      }),
    onSuccess: (result) => {
      trackEvent({
        event: 'install_sandbox_reprovision',
        user,
        status: 'ok',
        props: { orgId: org.id, installId: install.id },
      })
      addToast(
        <Toast heading="Sandbox reprovision started" theme="info">
          <Text>Reprovisioning <Badge variant="code" size="md">{install.name}</Badge> sandbox. This may take a few minutes.</Text>
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
        event: 'install_sandbox_reprovision',
        user,
        status: 'error',
        props: { orgId: org.id, installId: install.id, err: err?.error },
      })
      addToast(
        <Toast heading="Sandbox reprovision failed" theme="error">
          <Text>Unable to reprovision <Badge variant="code" size="md">{install.name}</Badge> sandbox.</Text>
        </Toast>
      )
    },
  })

  return (
    <ReprovisionSandboxModal
      installId={install?.id}
      isPending={isPending}
      error={error}
      onSubmit={({ selectedRole, skipComponents }) => {
        execute({
          body: {
            plan_only: false,
            skip_components: skipComponents,
            ...(selectedRole && { role: selectedRole }),
          },
        })
      }}
      onClose={() => removeModal(props.modalId)}
      {...props}
    />
  )
}

export const ReprovisionSandboxButton = ({
  ...props
}: IButtonAsButton) => {
  const { addModal } = useSurfaces()

  const modal = <ReprovisionSandboxModalContainer />
  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      {props?.isMenuButton ? null : <Icon variant="BoxArrowUpIcon" />}
      Reprovision sandbox
      {props?.isMenuButton ? <Icon variant="BoxArrowUpIcon" /> : null}
    </Button>
  )
}
