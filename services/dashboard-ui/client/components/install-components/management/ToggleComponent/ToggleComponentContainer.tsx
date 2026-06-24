import { useState } from 'react'
import { useNavigate } from 'react-router'
import { useMutation, useQueryClient } from '@tanstack/react-query'
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
import { toggleComponent } from '@/lib'
import type { TComponent } from '@/types'
import { ToggleComponentModal } from './ToggleComponent'

export const ToggleComponentModalContainer = ({
  component,
  enabling,
  ...props
}: Omit<IModal, 'onSubmit'> & {
  component: TComponent
  enabling: boolean
}) => {
  const navigate = useNavigate()
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const [selectedRole, setSelectedRole] = useState<string>('')

  const action = enabling ? 'Enabling' : 'Disabling'
  const pastAction = enabling ? 'enabled' : 'disabled'

  const { mutate: execute, isPending, error } = useMutation({
    mutationFn: (params: { role?: string }) =>
      toggleComponent({
        body: {
          enabled: enabling,
          plan_only: false,
          ...(params.role && { role: params.role }),
        },
        componentId: component.id,
        installId: install.id,
        orgId: org.id,
      }),
    onSuccess: (result) => {
      addToast(
        <Toast heading={`${action} component`} theme="info">
          <Text>
            {action} {component.name}. This may take a few minutes.
          </Text>
        </Toast>
      )
      queryClient.invalidateQueries({ queryKey: ['workflow-approvals'] })
      queryClient.invalidateQueries({ queryKey: ['active-workflows'] })
      queryClient.invalidateQueries({ queryKey: ['install-component'] })
      removeModal(props.modalId)
      const workflowId = result.data.workflow_id
      if (workflowId) {
        navigate(`/${org.id}/installs/${install.id}/workflows/${workflowId}`)
      } else {
        navigate(`/${org.id}/installs/${install.id}/workflows`)
      }
    },
    onError: (err: any) => {
      addToast(
        <Toast heading={`Component ${pastAction} failed`} theme="error">
          <Text>{err?.error || `Unable to ${pastAction} ${component.name}.`}</Text>
        </Toast>
      )
    },
  })

  return (
    <ToggleComponentModal
      component={component}
      enabling={enabling}
      isPending={isPending}
      error={error as any}
      onSubmit={({ role }) => {
        execute({ role: role || selectedRole })
      }}
      onClose={() => {
        removeModal(props.modalId)
      }}
      roleSelector={({ value, onChange }) => (
        <RoleSelector
          installId={install?.id}
          operationType={enabling ? 'deploy' : 'teardown'}
          principalType="component"
          principalId={component.id}
          value={value || selectedRole}
          onChange={(v) => {
            onChange(v)
            setSelectedRole(v)
          }}
          name="role"
        />
      )}
      {...props}
    />
  )
}

export const ToggleComponentButton = ({
  component,
  enabling,
  ...props
}: IButtonAsButton & {
  component: TComponent
  enabling: boolean
}) => {
  const { addModal } = useSurfaces()
  const modal = (
    <ToggleComponentModalContainer component={component} enabling={enabling} />
  )
  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      {props?.isMenuButton ? null : (
        <Icon variant={enabling ? 'ToggleRightIcon' : 'ToggleLeftIcon'} />
      )}
      {enabling ? 'Enable component' : 'Disable component'}
      {props?.isMenuButton ? (
        <Icon variant={enabling ? 'ToggleRightIcon' : 'ToggleLeftIcon'} />
      ) : null}
    </Button>
  )
}
