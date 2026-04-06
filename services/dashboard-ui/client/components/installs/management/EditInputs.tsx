import { useState, useRef } from 'react'
import { useNavigate } from 'react-router'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { Icon } from '@/components/common/Icon'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { UpdateInstallForm } from '@/components/installs/forms/UpdateInstallForm'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getAppConfig, updateInstall, updateInstallInputs } from '@/lib'
import type { TAppConfig } from '@/types'

interface IEditInputs {}

const ConfirmUpdateModal = ({
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

  if (!isInstallManagedByConfig) {
    onConfirm()
    return null
  }

  return (
    <Modal
      heading={
        <Text
          flex
          className="gap-4"
          variant="h3"
          weight="strong"
          theme="warn"
        >
          <Icon variant="Warning" size="24" />
          Override changes to this install?
        </Text>
      }
      primaryActionTrigger={{
        children: 'Confirm override',
        onClick: () => {
          onConfirm()
          removeModal(props.modalId)
        },
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-6">
        <div className="flex flex-col gap-4">
          <Text variant="body" weight="strong">
            You are about to update an Install managed by a Config file.
          </Text>
          <Text variant="body">
            If you proceed, the config file syncing will be disabled. Are you
            sure you want to continue?
          </Text>
        </div>

        <Banner theme="info">
          <Text variant="body">
            <strong>Tip:</strong> Use the management menu to enable Install
            Config syncing again.
          </Text>
        </Banner>

        <div className="flex gap-3 justify-end">
          <Button
            onClick={() => {
              onCancel()
              removeModal(props.modalId)
            }}
            variant="ghost"
          >
            Cancel
          </Button>
        </div>
      </div>
    </Modal>
  )
}

export const EditInputsFormModal = ({ ...props }: IEditInputs & IModal) => {
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
      const workflowId = result?.headers?.['x-nuon-install-workflow-id']
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

  const nestInputsUnderGroups = (
    groups: TAppConfig['input']['input_groups'],
    inputs: TAppConfig['input']['inputs']
  ) => {
    return groups
      ? groups.map((group) => ({
          ...group,
          app_inputs:
            inputs?.filter((input) => input.group_id === group.id) || [],
        }))
      : []
  }

  const handleFormSubmit = () => {
    if (formRef.current) {
      formRef.current.requestSubmit()
    }
  }

  const handleClose = () => {
    removeModal(props.modalId)
  }

  return (
    <Modal
      {...props}
      size="3/4"
      className="!max-h-[80vh]"
      childrenClassName="overflow-y-auto"
      heading={
        <Text
          flex
          className="gap-4"
          variant="h3"
          weight="strong"
        >
          <Icon variant="PencilSimpleLine" size="24" />
          Edit install inputs
        </Text>
      }
      primaryActionTrigger={
        !isLoading && !error && config
          ? {
              children: isSubmitting ? (
                <span className="flex items-center gap-2">
                  <Icon variant="Loading" />
                  Updating inputs
                </span>
              ) : (
                <span className="flex items-center gap-2">
                  <Icon variant="Cube" />
                  Update inputs
                </span>
              ),
              disabled: isSubmitting,
              onClick: handleFormSubmit,
              variant: 'primary',
            }
          : undefined
      }
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
            Deploy all dependents as well as the affected components.
          </Text>
        </div>
      }
      onClose={handleClose}
    >
      {isLoading ? (
        <div className="flex flex-col gap-8 max-w-3xl">
          <div className="flex flex-col gap-2">
            <Skeleton width="120px" height="20px" />
            <Skeleton width="200px" height="16px" />
          </div>

          <fieldset className="flex flex-col gap-6 border-t pt-6">
            <div className="flex flex-col gap-1 mb-6">
              <Skeleton width="280px" height="24px" />
              <Skeleton width="200px" height="16px" />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="flex flex-col gap-1">
                <Skeleton width="140px" height="16px" />
                <Skeleton width="180px" height="14px" />
              </div>
              <Skeleton width="100%" height="40px" />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="flex flex-col gap-1">
                <Skeleton width="120px" height="16px" />
                <Skeleton width="160px" height="14px" />
              </div>
              <Skeleton width="100%" height="40px" />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="flex flex-col gap-1">
                <Skeleton width="90px" height="16px" />
                <Skeleton width="220px" height="14px" />
              </div>
              <Skeleton width="100%" height="40px" />
            </div>
          </fieldset>

          <fieldset className="flex flex-col gap-6 border-t pt-6">
            <div className="flex flex-col gap-1 mb-6">
              <Skeleton width="320px" height="24px" />
              <Skeleton width="180px" height="16px" />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="flex flex-col gap-1">
                <Skeleton width="110px" height="16px" />
                <Skeleton width="140px" height="14px" />
              </div>
              <Skeleton width="100%" height="40px" />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="flex flex-col gap-1">
                <Skeleton width="80px" height="16px" />
                <Skeleton width="200px" height="14px" />
              </div>
              <Skeleton width="100%" height="40px" />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div />
              <div className="flex items-center gap-2 ml-1">
                <Skeleton width="16px" height="16px" />
                <Skeleton width="130px" height="16px" />
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div />
              <div className="flex items-center gap-2 ml-1">
                <Skeleton width="16px" height="16px" />
                <Skeleton width="110px" height="16px" />
              </div>
            </div>
          </fieldset>

          <fieldset className="flex flex-col gap-4 border-t pt-6">
            <div className="flex flex-col gap-1 mb-4">
              <Skeleton width="200px" height="24px" />
              <Skeleton width="400px" height="16px" />
            </div>

            <div className="flex gap-6">
              <div className="flex items-center gap-2">
                <Skeleton width="16px" height="16px" />
                <Skeleton width="160px" height="16px" />
              </div>
              <div className="flex items-center gap-2">
                <Skeleton width="16px" height="16px" />
                <Skeleton width="140px" height="16px" />
              </div>
            </div>
          </fieldset>
        </div>
      ) : error?.error ? (
        <Banner theme="error">
          {error?.error || 'Unable to load app configuration'}
        </Banner>
      ) : (
        <UpdateInstallForm
          ref={formRef}
          install={install}
          inputConfig={{
            ...config?.input,
            input_groups: nestInputsUnderGroups(
              config?.input?.input_groups,
              config?.input?.inputs
            ),
          }}
          onSubmit={(formData) => mutateAsync(formData)}
          onCancel={() => {
            removeModal(props.modalId)
          }}
          onRegisterClearDraft={(fn) => {
            clearDraftRef.current = fn
          }}
          selectedRole={selectedRole}
          onRoleChange={setSelectedRole}
        />
      )}
    </Modal>
  )
}

export const EditInputsButton = ({
  ...props
}: IEditInputs & IButtonAsButton) => {
  const { install } = useInstall()
  const { addModal } = useSurfaces()

  const showConfirmModal = () => {
    const confirmModal = (
      <ConfirmUpdateModal
        onConfirm={() => {
          const editModal = <EditInputsFormModal />
          addModal(editModal)
        }}
        onCancel={() => {}}
      />
    )
    addModal(confirmModal)
  }

  const showEditModal = () => {
    const editModal = <EditInputsFormModal />
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
