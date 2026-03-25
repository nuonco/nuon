import { useNavigate } from 'react-router'
import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useAuth } from '@/hooks/use-auth'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Input } from '@/components/common/form/Input'
import { RoleSelector } from '@/components/roles/RoleSelector'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { deprovisionSandbox, getAppConfig } from '@/lib'
import { trackEvent } from '@/lib/segment-analytics'

export const DeprovisionSandboxButton = ({
  ...props
}: IButtonAsButton) => {
  const { addModal } = useSurfaces()

  const modal = <DeprovisionSandboxModal />
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

export const DeprovisionSandboxModal = ({
  ...props
}: IModal) => {
  const navigate = useNavigate()
  const { user } = useAuth()
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const { data: appConfig } = useQuery({
    queryKey: ['app-config', org?.id, install?.app_id, install?.app_config_id, 'recurse'],
    queryFn: () =>
      getAppConfig({
        orgId: org.id,
        appId: install.app_id,
        appConfigId: install.app_config_id,
        recurse: true,
      }),
    enabled: !!org?.id && !!install?.app_config_id,
  })

  const sandboxOperationRoles = appConfig?.sandbox?.operation_roles
  const defaultRole = sandboxOperationRoles?.deprovision
    ?.replace('{{.nuon.install.id}}', install.id)
  const mappedRoleNames = sandboxOperationRoles
    ? Object.values(sandboxOperationRoles).map((r) => r.replace('{{.nuon.install.id}}', install.id))
    : undefined

  const [confirm, setConfirm] = useState<string>('')
  const [selectedRole, setSelectedRole] = useState<string>('')

  const { mutate: execute, isPending: isLoading, error } = useMutation({
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
      const workflowId = result?.headers?.['x-nuon-install-workflow-id']
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

  const handleDeprovision = () => {
    execute({
      body: {
        plan_only: false,
        error_behavior: 'abort',
        ...(selectedRole && { role: selectedRole }),
      },
    })
  }

  const handleClose = () => {
    removeModal(props.modalId)
  }

  return (
    <Modal
      className="!max-w-xl"
      heading="Deprovision install sandbox"
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" />
            Deprovisioning
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="BoxArrowDown" />
            Deprovision sandbox
          </span>
        ),
        disabled: confirm !== 'deprovision' || isLoading,
        onClick: handleDeprovision,
        variant: 'danger' as const,
      }}
      onClose={handleClose}
      {...props}
    >
      <div className="flex flex-col gap-8 mb-12">
        {(error as any)?.error ? (
          <Banner theme="error">
            {(error as any)?.error || 'Unable to kickoff sandbox deprovision'}
          </Banner>
        ) : null}

        <span className="flex flex-col gap-1">
          <Text variant="h3" weight="strong">
            Are you sure you want to deprovision {install?.name} sandbox?
          </Text>
          <Text
            className="text-cool-grey-600 dark:text-white/70"
            variant="subtext"
          >
            Deprovisioning a sandbox will remove it from the cloud account.
          </Text>
        </span>

        <div className="flex flex-col gap-2">
          <Text variant="body">
            This will create a workflow that attempts to:
          </Text>

          <ul className="flex flex-col gap-1 list-disc pl-4">
            <li className="text-sm">Teardown the install sandbox</li>
          </ul>
        </div>

        <div className="w-full">
          <label className="flex flex-col gap-1 w-full">
            <Text variant="base" weight="strong">
              To verify, type{' '}
              <span className="text-red-800 dark:text-red-500">
                deprovision
              </span>{' '}
              below.
            </Text>
            <Input
              placeholder="deprovision"
              className="w-full"
              type="text"
              value={confirm}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                setConfirm(e?.currentTarget?.value)
              }}
            />
          </label>
        </div>

        <RoleSelector
          installId={install?.id}
          operationType="deprovision"
          principalType='sandbox'
          value={selectedRole}
          onChange={setSelectedRole}
          name="role"
          defaultRoleName={defaultRole}
          mappedRoleNames={mappedRoleNames}
        />
      </div>
    </Modal>
  )
}
