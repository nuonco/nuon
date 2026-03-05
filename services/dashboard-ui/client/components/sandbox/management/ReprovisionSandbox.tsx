import { useNavigate } from 'react-router'
import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { useAuth } from '@/hooks/use-auth'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { RoleSelector } from '@/components/common/form/RoleSelector'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { reprovisionSandbox } from '@/lib'
import { trackEvent } from '@/lib/segment-analytics'

export const ReprovisionSandboxButton = ({
  ...props
}: IButtonAsButton) => {
  const { addModal } = useSurfaces()

  const modal = <ReprovisionSandboxModal />
  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      {props?.isMenuButton ? null : <Icon variant="BoxArrowUp" />}
      Reprovision sandbox
      {props?.isMenuButton ? <Icon variant="BoxArrowUp" /> : null}
    </Button>
  )
}

export const ReprovisionSandboxModal = ({
  ...props
}: IModal) => {
  const navigate = useNavigate()
  const { user } = useAuth()
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const [selectedRole, setSelectedRole] = useState<string>('')

  const { mutate: execute, isPending: isLoading, error } = useMutation({
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
        <Toast heading="Reprovision initiated" theme="success">
          <Text>Sandbox reprovision workflow has been started successfully.</Text>
        </Toast>
      )
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
        event: 'install_sandbox_reprovision',
        user,
        status: 'error',
        props: { orgId: org.id, installId: install.id, err: err?.error },
      })
      addToast(
        <Toast heading="Sandbox reprovision failed" theme="error">
          <Text>Failed to start sandbox reprovision. Please try again.</Text>
        </Toast>
      )
    },
  })

  const handleReprovision = () => {
    execute({
      body: {
        plan_only: false,
        ...(selectedRole && { role: selectedRole }),
      },
    })
  }

  const handleClose = () => {
    removeModal(props.modalId)
  }

  return (
    <Modal
      heading="Reprovision sandbox?"
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" />
            Reprovisioning
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="BoxArrowUp" />
            Reprovision sandbox
          </span>
        ),
        disabled: isLoading,
        onClick: handleReprovision,
        variant: 'primary' as const,
      }}
      onClose={handleClose}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {(error as any)?.error ? (
          <Banner theme="error">
            {(error as any)?.error || 'Unable to reprovision sandbox.'}
          </Banner>
        ) : null}

        <Text variant="body" className="leading-relaxed">
          Are you sure you want to reprovision this sandbox?
        </Text>

        <RoleSelector
          installId={install?.id}
          operationType="reprovision"
          principalType="sandbox"
          value={selectedRole}
          onChange={(e) => setSelectedRole(e.target.value)}
          name="role"
        />
      </div>
    </Modal>
  )
}
