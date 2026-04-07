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
import { deployComponent } from '@/lib'
import { trackEvent } from '@/lib/segment-analytics'
import type { TComponent } from '@/types'
import { BuildSelect } from '../BuildSelect'
import { DeployComponentModal } from './DeployComponent'

export const DeployComponentModalContainer = ({
  component,
  currentBuildId,
  currentDeployStatus,
  ...props
}: Omit<IModal, 'onSubmit'> & {
  component: TComponent
  currentBuildId?: string
  currentDeployStatus?: string
}) => {
  const navigate = useNavigate()
  const { user } = useAuth()
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const { mutate: execute, isPending, error } = useMutation({
    mutationFn: (params: { body: Parameters<typeof deployComponent>[0]['body'] }) =>
      deployComponent({
        body: params.body,
        installId: install.id,
        orgId: org.id,
      }),
    onSuccess: (result) => {
      trackEvent({
        event: 'component_deploy',
        status: 'ok',
        user,
        props: {
          orgId: org.id,
          installId: install.id,
          componentId: component.id,
        },
      })
      addToast(
        <Toast heading={`${component.name} deploy started`} theme="success">
          <Text>Deploy for {component.name} was started.</Text>
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
        event: 'component_deploy',
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
        <Toast heading="Component deploy failed" theme="error">
          <Text>Unable to deploy {component.name} component.</Text>
        </Toast>
      )
    },
  })

  return (
    <DeployComponentModal
      component={component}
      currentBuildId={currentBuildId}
      currentDeployStatus={currentDeployStatus}
      installId={install.id}
      isPending={isPending}
      error={error as any}
      onSubmit={({ buildId, deployDependents, role }) => {
        execute({
          body: {
            build_id: buildId,
            deploy_dependents: deployDependents,
            plan_only: false,
            ...(role && { role }),
          },
        })
      }}
      onClose={() => removeModal(props.modalId)}
      buildSelect={({ selectedBuildId, onSelectBuild, onClose }) => (
        <BuildSelect
          componentId={component.id}
          componentType={component.type}
          selectedBuildId={selectedBuildId}
          currentBuildId={currentBuildId}
          currentDeployStatus={currentDeployStatus}
          onSelectBuild={onSelectBuild}
          onClose={onClose}
        />
      )}
      roleSelector={({ value, onChange }) => (
        <RoleSelector
          installId={install?.id}
          operationType="deploy"
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

export const DeployComponentButton = ({
  component,
  currentBuildId,
  currentDeployStatus,
  ...props
}: IButtonAsButton & {
  component: TComponent
  currentBuildId?: string
  currentDeployStatus?: string
}) => {
  const { addModal } = useSurfaces()
  const modal = (
    <DeployComponentModalContainer
      component={component}
      currentBuildId={currentBuildId}
      currentDeployStatus={currentDeployStatus}
    />
  )
  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      {props?.isMenuButton ? null : <Icon variant="CloudArrowUp" />}
      Deploy component
      {props?.isMenuButton ? <Icon variant="CloudArrowUp" /> : null}
    </Button>
  )
}
