import { useState, useRef } from 'react'
import { useNavigate } from 'react-router'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getAppConfig, updateInstall, updateInstallInputs } from '@/lib'
import { ConfirmUpdateModal, EditInputsFormModal } from './EditInputs'

interface IEditInputs {}

export const ConfirmUpdateModalContainer = ({
  onConfirm,
  onCancel,
  ...props
}: {
  onConfirm: () => void
  onCancel: () => void
} & IModal) => {
  const { install } = useInstall()
  const { removeModal } = useSurfaces()

  const isInstallManagedByConfig =
    install?.metadata?.managed_by === 'nuon/cli/install-config'

  return (
    <ConfirmUpdateModal
      isInstallManagedByConfig={isInstallManagedByConfig}
      onConfirm={() => {
        onConfirm()
        removeModal(props.modalId)
      }}
      onCancel={() => {
        onCancel()
        removeModal(props.modalId)
      }}
      {...props}
    />
  )
}

export const EditInputsFormModalContainer = ({ ...props }: IEditInputs & Omit<IModal, 'onSubmit'>) => {
  const navigate = useNavigate()
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const formRef = useRef<HTMLFormElement>(null)
  const clearDraftRef = useRef<(() => void) | null>(null)
  const [selectedRole, setSelectedRole] = useState<string>('')
  const [deployDependents, setDeployDependents] = useState(true)

  const {
    data: config,
    isLoading,
    error,
  } = useQuery({
    queryKey: ['app-config', org.id, install?.app_id, install?.app_config_id],
    queryFn: () =>
      getAppConfig({
        orgId: org.id,
        appId: install.app_id,
        appConfigId: install.app_config_id,
        recurse: true,
      }),
    enabled: !!install?.app_id && !!install?.app_config_id,
  })

  const { mutateAsync, isPending: isSubmitting, error: actionError } = useMutation({
    mutationFn: async (formData: FormData) => {
      if (install?.metadata?.managed_by === 'nuon/cli/install-config') {
        await updateInstall({
          body: { metadata: { managed_by: 'nuon/dashboard' } },
          installId: install.id,
          orgId: org.id,
        })
      }

      const formDataObj = Object.fromEntries(formData)
      const inputs = Object.keys(formDataObj).reduce(
        (acc, key) => {
          if (key.includes('inputs:')) {
            let value: any = formDataObj[key]
            if (value === 'on' || value === 'off') {
              value = Boolean(value === 'on').toString()
            }
            acc[key.replace('inputs:', '')] = value
          }
          return acc
        },
        {} as Record<string, any>
      )

      return updateInstallInputs({
        installId: install.id,
        orgId: org.id,
        body: {
          inputs,
          deploy_dependents: deployDependents,
          ...(selectedRole && { role: selectedRole }),
        },
      })
    },
    onSuccess: (result) => {
      addToast(
        <Toast heading="Inputs updated" theme="success">
          <Text>Install inputs updated successfully!</Text>
        </Toast>
      )
      queryClient.invalidateQueries({ queryKey: ['workflow-approvals'] })
      queryClient.invalidateQueries({ queryKey: ['active-workflows'] })
      removeModal(props.modalId)
      const workflowId = result.data.workflow_id
      if (workflowId) {
        navigate(`/${org.id}/installs/${install?.id}/workflows/${workflowId}`)
      } else {
        navigate(`/${org.id}/installs/${install?.id}/workflows`)
      }
    },
    onError: (error) => {
      addToast(
        <Toast heading="Update failed" theme="error">
          <Text>Unable to update install inputs.</Text>
        </Toast>
      )
    },
  })

  const handleFormSubmit = () => {
    if (formRef.current) {
      formRef.current.requestSubmit()
    }
  }

  const handleClose = () => {
    removeModal(props.modalId)
  }

  return (
    <EditInputsFormModal
      install={install}
      config={config}
      isLoading={isLoading}
      error={error}
      isSubmitting={isSubmitting}
      actionError={actionError}
      onFormSubmit={handleFormSubmit}
      onClose={handleClose}
      formRef={formRef}
      clearDraftRef={clearDraftRef}
      selectedRole={selectedRole}
      onRoleChange={setSelectedRole}
      deployDependents={deployDependents}
      onDeployDependentsChange={setDeployDependents}
      onMutate={(formData) => mutateAsync(formData)}
      {...props}
    />
  )
}

export const EditInputsButton = ({
  ...props
}: IEditInputs & IButtonAsButton) => {
  const { install } = useInstall()
  const { addModal } = useSurfaces()

  const showConfirmModal = () => {
    const confirmModal = (
      <ConfirmUpdateModalContainer
        onConfirm={() => {
          const editModal = <EditInputsFormModalContainer />
          addModal(editModal)
        }}
        onCancel={() => {}}
      />
    )
    addModal(confirmModal)
  }

  const showEditModal = () => {
    const editModal = <EditInputsFormModalContainer />
    addModal(editModal)
  }

  const handleClick = () => {
    const isInstallManagedByConfig =
      install?.metadata?.managed_by === 'nuon/cli/install-config'

    if (isInstallManagedByConfig) {
      showConfirmModal()
    } else {
      showEditModal()
    }
  }

  return (
    <Button
      className="text-sm !font-medium !py-2 !px-3 h-[36px] flex items-center gap-3 w-full"
      variant="ghost"
      onClick={handleClick}
      {...props}
    >
      Edit inputs
      <Icon variant="PencilSimpleLine" />
    </Button>
  )
}
