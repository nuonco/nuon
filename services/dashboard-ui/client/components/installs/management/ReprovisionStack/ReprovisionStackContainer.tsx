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
import { useInstallLink } from '@/hooks/use-install-path'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { reprovisionStack } from '@/lib'
import { trackEvent } from '@/lib/posthog-analytics'
import { ReprovisionStackModal } from './ReprovisionStack'

export const ReprovisionStackModalContainer = ({
  onSubmit: _onSubmit,
  ...props
}: IModal) => {
  const navigate = useNavigate()
  const { user } = useAuth()
  const { org } = useOrg()
  const { install } = useInstall()
  const installLink = useInstallLink()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const { mutate: execute, isPending, error } = useMutation({
    mutationFn: (params: { body: Parameters<typeof reprovisionStack>[0]['body'] }) =>
      reprovisionStack({
        body: params.body,
        installId: install.id,
        orgId: org.id,
      }),
    onSuccess: (result) => {
      trackEvent({
        event: 'install_stack_reprovision',
        user,
        status: 'ok',
        props: { orgId: org.id, installId: install.id },
      })
      addToast(
        <Toast heading="Stack reprovision started" theme="info">
          <Text>Reprovisioning the stack for <Badge variant="code" size="md">{install.name}</Badge>. This may take a few minutes.</Text>
        </Toast>
      )
      queryClient.invalidateQueries({ queryKey: ['workflow-approvals'] })
      queryClient.invalidateQueries({ queryKey: ['active-workflows'] })
      queryClient.invalidateQueries({ queryKey: ['install-stack'] })
      removeModal(props.modalId)
      const workflowId = result.data.workflow_id
      if (workflowId) {
        navigate(installLink({ installId: install.id, appId: install.app_id, suffix: `/workflows/${workflowId}` }))
      } else {
        navigate(installLink({ installId: install.id, appId: install.app_id, suffix: `/workflows` }))
      }
    },
    onError: (err: any) => {
      trackEvent({
        event: 'install_stack_reprovision',
        user,
        status: 'error',
        props: { orgId: org.id, installId: install.id, err: err?.error },
      })
      addToast(
        <Toast heading="Stack reprovision failed" theme="error">
          <Text>Unable to reprovision the stack for <Badge variant="code" size="md">{install.name}</Badge>.</Text>
        </Toast>
      )
    },
  })

  return (
    <ReprovisionStackModal
      installName={install?.name}
      isPending={isPending}
      error={error}
      onSubmit={() => {
        execute({
          body: {
            plan_only: false,
          },
        })
      }}
      onClose={() => removeModal(props.modalId)}
      {...props}
    />
  )
}

export const ReprovisionStackButton = ({ ...props }: IButtonAsButton) => {
  const { addModal } = useSurfaces()

  const modal = <ReprovisionStackModalContainer />
  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      {props?.isMenuButton ? null : <Icon variant="StackPlusIcon" />}
      Reprovision stack
      {props?.isMenuButton ? <Icon variant="StackPlusIcon" /> : null}
    </Button>
  )
}
