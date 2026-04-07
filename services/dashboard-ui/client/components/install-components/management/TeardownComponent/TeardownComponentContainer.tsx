import { useNavigate } from 'react-router'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useAuth } from '@/hooks/use-auth'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { RoleSelector } from '@/components/roles/RoleSelector'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { teardownComponent } from '@/lib'
import { trackEvent } from '@/lib/segment-analytics'
import type { TComponent } from '@/types'
import { TeardownComponentModal } from './TeardownComponent'

export const TeardownComponentModalContainer = ({
  component,
  ...props
}: Omit<IModal, 'onSubmit'> & {
  component: TComponent
}) => {
  const navigate = useNavigate()
  const { user } = useAuth()
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const { mutate: execute, isPending, error } = useMutation({
    mutationFn: (params: { body: Parameters<typeof teardownComponent>[0]['body'] }) =>
      teardownComponent({
        body: params.body,
        componentId: component.id,
        installId: install.id,
        orgId: org.id,
      }),
    onSuccess: (result) => {
      trackEvent({
        event: 'component_teardown',
        status: 'ok',
        user,
        props: {
          orgId: org.id,
          installId: install.id,
          componentId: component.id,
        },
      })
      addToast(
        <Toast heading={`${component.name} teardown started`} theme="success">
          <Text>Teardown for {component.name} was started.</Text>
        </Toast>
      )
      queryClient.invalidateQueries({ queryKey: ['workflow-approvals'] })
      queryClient.invalidateQueries({ queryKey: ['active-workflows'] })
      removeModal(props.modalId)
      const workflowId = result?.headers?.['x-nuon-install-workflow-id']
      if (workflowId) {
        navigate(`/${org.id}/installs/${install.id}/workflows/${workflowId}`)
      } else {
        navigate(`/${org.id}/installs/${install.id}/workflows`)
      }
    },
    onError: (err: any) => {
      trackEvent({
        event: 'component_teardown',
        status: 'error',
        user,
        props: {
          orgId: org.id,
          installId: install.id,
          componentId: component.id,
          err: err?.error,
        },
      })
      addToast(
        <Toast heading="Component teardown failed" theme="error">
          <Text>Unable to teardown {component.name} component.</Text>
        </Toast>
      )
    },
  })

  return (
    <TeardownComponentModal
      component={component}
      isPending={isPending}
      error={error as any}
      onSubmit={({ role }) => {
        execute({
          body: {
            plan_only: false,
            error_behavior: 'continue',
            ...(role && { role }),
          },
        })
      }}
      onClose={() => {
        removeModal(props.modalId)
      }}
      roleSelector={({ value, onChange }) => (
        <RoleSelector
          installId={install?.id}
          operationType="teardown"
          principalType="component"
          principalId={component.id}
          value={value}
          onChange={onChange}
          name="role"
        />
      )}
      {...props}
    />
  )
}

export const TeardownComponentButton = ({
  component,
  ...props
}: IButtonAsButton & {
  component: TComponent
}) => {
  const { addModal } = useSurfaces()
  const modal = <TeardownComponentModalContainer component={component} />
  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
      className="!text-red-800 dark:!text-red-500"
    >
      {props?.isMenuButton ? null : <Icon variant="CloudArrowDownIcon" />}
      Teardown component
      {props?.isMenuButton ? <Icon variant="CloudArrowDownIcon" /> : null}
    </Button>
  )
}
