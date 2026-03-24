import { useNavigate } from 'react-router'
import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useAuth } from '@/hooks/use-auth'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { RoleSelector } from '@/components/roles/RoleSelector'
import { Toast } from '@/components/surfaces/Toast'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { teardownComponent } from '@/lib'
import { trackEvent } from '@/lib/segment-analytics'
import type { TComponent } from '@/types'

export const TeardownComponentButton = ({
  component,
  ...props
}: IButtonAsButton & {
  component: TComponent
}) => {
  const { addModal } = useSurfaces()
  const modal = <TeardownComponentModal component={component} />
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

export const TeardownComponentModal = ({
  component,
  ...props
}: IModal & {
  component: TComponent
}) => {
  const navigate = useNavigate()
  const { user } = useAuth()
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const [confirmName, setConfirmName] = useState('')
  const [selectedRole, setSelectedRole] = useState<string>('')

  const { mutate: execute, isPending: isLoading, error } = useMutation({
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

  const handleClose = () => {
    setConfirmName('')
    removeModal(props.modalId)
  }

  const handleTeardown = () => {
    execute({
      body: {
        plan_only: false,
        error_behavior: 'continue',
        ...(selectedRole && { role: selectedRole }),
      },
    })
  }

  const isConfirmValid = confirmName === component.name
  const canTeardown = isConfirmValid && !isLoading

  return (
    <Modal
      heading={
        <div className="flex flex-col gap-2">
          <Text
            className="inline-flex gap-4 items-center"
            variant="h3"
            weight="strong"
            theme="error"
          >
            <Icon variant="Warning" size="24" />
            Teardown {component.name} component
          </Text>
          <Text
            variant="body"
            className="text-cool-grey-600 dark:text-cool-grey-400"
          >
            This will remove all running deployments for this component
          </Text>
        </div>
      }
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" />
            Tearing down
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="CloudArrowDownIcon" />
            Teardown component
          </span>
        ),
        disabled: !canTeardown,
        onClick: handleTeardown,
        variant: 'danger' as const,
      }}
      onClose={handleClose}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {(error as any)?.error ? (
          <Banner theme="error">
            {(error as any)?.error || 'Unable to teardown component'}
          </Banner>
        ) : null}

        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-2">
            <Text variant="base" weight="strong">
              Are you sure you want to teardown {component.name}?
            </Text>
            <Text variant="body" theme="neutral">
              Tearing down a component will remove all of its running deployments from the cloud account.
            </Text>
          </div>

          <div className="flex flex-col gap-3">
            <Text variant="body">
              This will create a workflow that attempts to:
            </Text>
            <ul className="flex flex-col gap-1 list-disc pl-6 text-sm text-cool-grey-700 dark:text-cool-grey-300">
              <li>Remove all infrastructure provisioned by this component</li>
              <li>Clean up all associated resources and dependencies</li>
            </ul>
          </div>

          <div className="flex flex-col gap-2">
            <Text variant="body">
              To verify, type{' '}
              <span className="font-mono font-medium text-red-800 dark:text-red-400 bg-red-50 dark:bg-red-900/20 px-1 py-0.5 rounded">
                {component.name}
              </span>{' '}
              below.
            </Text>
            <Input
              id="confirm-component-name"
              placeholder="component name"
              type="text"
              value={confirmName}
              onChange={(e) => setConfirmName(e.target.value)}
              error={confirmName.length > 0 && !isConfirmValid}
              errorMessage={confirmName.length > 0 && !isConfirmValid ? "Component name doesn't match" : undefined}
            />
          </div>

          <Banner theme="warn">
            <Text variant="body">
              <strong>Important:</strong> This action cannot be undone. All infrastructure provisioned by this component will be permanently destroyed.
            </Text>
          </Banner>

          <RoleSelector
            installId={install?.id}
            operationType="teardown"
            principalType="component"
            value={selectedRole}
            onChange={setSelectedRole}
            name="role"
          />
        </div>
      </div>
    </Modal>
  )
}
