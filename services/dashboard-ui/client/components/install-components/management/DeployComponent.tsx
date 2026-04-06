import { useNavigate } from 'react-router'
import { useEffect, useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useAuth } from '@/hooks/use-auth'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { RoleSelector } from '@/components/roles/RoleSelector'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { deployComponent } from '@/lib'
import { trackEvent } from '@/lib/segment-analytics'
import type { TComponent } from '@/types'
import { BuildSelect } from './BuildSelect'

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
    <DeployComponentModal
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

export const DeployComponentModal = ({
  component,
  currentBuildId,
  currentDeployStatus,
  ...props
}: IModal & {
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

  const [buildId, setBuildId] = useState<string>()
  const [deployDependents, setDeployDependents] = useState(false)
  const [selectedRole, setSelectedRole] = useState<string>('')

  const { mutate: execute, isPending: isLoading, error, data: deploy } = useMutation({
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
          buildId,
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
          buildId,
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

  const handleClose = () => {
    setBuildId(undefined)
    removeModal(props.modalId)
  }

  const handleBuildSelect = (selectedBuildId: string) => {
    setBuildId(selectedBuildId)
  }

  const isDeployDisabled = !buildId || isLoading

  return (
    <Modal
      heading={
        <div className="flex flex-col gap-2">
          <Text
            flex
            className="gap-4"
            variant="h3"
            weight="strong"
            theme="info"
          >
            <Icon variant="CloudArrowUp" size="24" />
            Deploy {component.name} component
          </Text>
          <Text
            variant="body"
            className="text-cool-grey-600 dark:text-cool-grey-400"
          >
            Select a build to deploy to your install
          </Text>
        </div>
      }
      size="half"
      className="!max-h-[80vh]"
      childrenClassName="flex-auto overflow-y-auto"
      onClose={handleClose}
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" />
            Deploying build
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="CloudArrowUp" />
            Deploy build
          </span>
        ),
        disabled: isDeployDisabled,
        onClick: () => {
          execute({
            body: {
              build_id: buildId!,
              deploy_dependents: deployDependents,
              plan_only: false,
              ...(selectedRole && { role: selectedRole }),
            },
          })
        },
        variant: 'primary' as const,
      }}
      footerActions={
        <div className="flex flex-col gap-1 pl-4">
          <CheckboxInput
            checked={deployDependents}
            onChange={(e) => setDeployDependents(e.target.checked)}
            labelProps={{
              className:
                'hover:!bg-transparent focus:!bg-transparent active:!bg-transparent !px-0 !py-1 gap-4 max-w-none',
              labelText: 'Deploy dependents',
              labelTextProps: { variant: 'base', weight: 'stronger' },
            }}
          />
          <Text variant="subtext" theme="neutral" className="ml-8 leading-none">
            Deploy all dependents as well as the selected build.
          </Text>
        </div>
      }
      {...props}
    >
      <div className="flex flex-col gap-6">
        {(error as any)?.error ? (
          <Banner theme="error">
            {(error as any)?.error || 'Unable to deploy component'}
          </Banner>
        ) : null}

        <BuildSelect
          componentId={component.id}
          componentType={component.type}
          selectedBuildId={buildId}
          currentBuildId={currentBuildId}
          currentDeployStatus={currentDeployStatus}
          onSelectBuild={handleBuildSelect}
          onClose={handleClose}
        />

        <RoleSelector
          installId={install?.id}
          operationType="deploy"
          principalType="component"
          principalId={component.id}
          value={selectedRole}
          onChange={setSelectedRole}
          name="role"
        />
      </div>
    </Modal>
  )
}
