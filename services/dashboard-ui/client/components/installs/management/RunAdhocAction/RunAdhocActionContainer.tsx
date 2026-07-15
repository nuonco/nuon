import { useState } from 'react'
import { useNavigate } from 'react-router'
import { useMutation } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { ResumeDraftModal } from '@/components/installs/forms/shared/ResumeDraftModal'
import { RoleSelector } from '@/components/roles/RoleSelector'
import type { IModal } from '@/components/surfaces/Modal'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { runAdhocAction, type TRunAdhocActionBody } from '@/lib'
import { RunAdhocActionModal } from './RunAdhocAction'

interface IRunAdhocAction {
  initialValues?: TRunAdhocActionBody
}

export const RunAdhocActionModalContainer = ({
  initialValues,
  onSubmit: _onSubmit,
  ...props
}: IRunAdhocAction & IModal) => {
  const navigate = useNavigate()
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal, addModal } = useSurfaces()
  const { addToast } = useToast()
  const [selectedRole, setSelectedRole] = useState<string>(
    initialValues?.role || ''
  )

  const {
    mutate,
    isPending: isLoading,
    error,
  } = useMutation({
    mutationFn: (body: TRunAdhocActionBody) =>
      runAdhocAction({
        body: {
          ...body,
          ...(selectedRole && { role: selectedRole }),
        },
        installId: install.id,
        orgId: org.id,
      }),
    onSuccess: (result) => {
      addToast(
        <Toast heading="Adhoc action started" theme="success">
          <Text>Adhoc action is running.</Text>
        </Toast>
      )
      removeModal(props.modalId)
      const workflowId = result.data.workflow_id
      if (workflowId) {
        navigate(`/${org.id}/installs/${install.id}/workflows/${workflowId}`)
      } else {
        navigate(`/${org.id}/installs/${install.id}/workflows`)
      }
    },
    onError: (error) => {
      addToast(
        <Toast heading="Adhoc action failed" theme="error">
          <Text>Unable to run adhoc action.</Text>
        </Toast>
      )
    },
  })

  const handleDraftResume = (
    onResume: () => void,
    onStartFresh: () => void,
    onClose: () => void,
    draftTimestamp: string
  ) => {
    let modalId: string
    const modal = (
      <ResumeDraftModal
        draftTimestamp={draftTimestamp}
        onResume={() => {
          onResume()
          removeModal(modalId)
        }}
        onStartFresh={() => {
          onStartFresh()
          removeModal(modalId)
        }}
        onClose={() => {
          onClose()
          removeModal(modalId)
        }}
      />
    )
    modalId = addModal(modal)
  }

  return (
    <RunAdhocActionModal
      installId={install.id}
      initialValues={initialValues}
      isPending={isLoading}
      error={error}
      onSubmit={(body) => mutate(body)}
      onDraftResume={handleDraftResume}
      roleSelector={
        <RoleSelector
          installId={install.id}
          operationType="trigger"
          principalType="action"
          value={selectedRole}
          onChange={setSelectedRole}
        />
      }
      {...props}
    />
  )
}

export const RunAdhocActionButton = ({
  initialValues,
  children,
  ...props
}: IRunAdhocAction & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <RunAdhocActionModalContainer initialValues={initialValues} />

  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      {children ?? (
        <>
          Run adhoc action
          <Icon variant="TerminalWindowIcon" />
        </>
      )}
    </Button>
  )
}
