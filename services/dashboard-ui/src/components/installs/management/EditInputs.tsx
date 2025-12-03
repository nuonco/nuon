'use client'

import { useState, useRef } from 'react'
import { usePathname, useRouter } from 'next/navigation'
import { updateInstallInputs } from '@/actions/installs/update-install-inputs'
import { updateInstall } from '@/actions/installs/update-install'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { UpdateInstallForm } from '@/components/installs/forms/UpdateInstallForm'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useQuery } from '@/hooks/use-query'
import { useSurfaces } from '@/hooks/use-surfaces'
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
    // Auto-proceed if not config managed
    onConfirm()
    return null
  }

  return (
    <Modal
      heading={
        <Text
          className="inline-flex gap-4 items-center"
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
            If you proceed, the config file syncing will be disabled. Are you sure you want to continue?
          </Text>
        </div>
        
        <Banner theme="info">
          <Text variant="body">
            <strong>Tip:</strong> Use the management menu to enable Install Config syncing again.
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

const EditInputsFormModal = ({ ...props }: IEditInputs & IModal) => {
  const path = usePathname()
  const router = useRouter()
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal } = useSurfaces()
  const formRef = useRef<HTMLFormElement>(null)

  const {
    data: config,
    isLoading,
    error,
  } = useQuery<TAppConfig>({
    path: `/api/orgs/${org.id}/apps/${install?.app_id}/configs/${install?.app_config_id}?recurse=true`,
  })

  const nestInputsUnderGroups = (
    groups: TAppConfig['input']['input_groups'],
    inputs: TAppConfig['input']['inputs']
  ) => {
    return groups
      ? groups.map((group) => ({
          ...group,
          app_inputs: inputs?.filter((input) => input.group_id === group.id) || [],
        }))
      : []
  }

  const handleSuccess = ({ headers }: { headers?: Record<string, string> }) => {
    const workflowId = headers?.['x-nuon-install-workflow-id']
    if (workflowId) {
      router.push(`/${org.id}/installs/${install?.id}/workflows/${workflowId}`)
    }
    removeModal(props.modalId)
  }

  const handleFormDataSubmit = async (formData: FormData) => {
    // If install is managed by config, switch it to dashboard managed
    if (install?.metadata?.managed_by === 'nuon/cli/install-config') {
      await updateInstall({
        body: { metadata: { managed_by: 'nuon/dashboard' } },
        installId: install.id,
        orgId: org.id,
        path,
      })
    }

    return updateInstallInputs({
      installId: install.id,
      orgId: org.id,
      formData,
      path,
    })
  }


  const handleFormSubmit = () => {
    if (formRef.current) {
      formRef.current.requestSubmit()
    }
  }

  return (
    <Modal
      className="!max-w-5xl"
      childrenClassName="!max-h-[80vh] overflow-y-auto"
      heading={
        <Text
          className="inline-flex gap-4 items-center"
          variant="h3"
          weight="strong"
        >
          <Icon variant="PencilSimpleLine" size="24" />
          Edit install inputs
        </Text>
      }
      primaryActionTrigger={
        !isLoading && !error && config ? {
          children: (
            <span className="flex items-center gap-2">
              <Icon variant="Cube" />
              Update inputs
            </span>
          ),
          onClick: handleFormSubmit,
          variant: 'primary',
        } : undefined
      }
      {...props}
    >
      {isLoading ? (
        <div className="flex flex-col gap-8 max-w-3xl">
          {/* Install name section */}
          <div className="flex flex-col gap-2">
            <Skeleton width="120px" height="20px" />
            <Skeleton width="200px" height="16px" />
          </div>

          {/* First input group */}
          <fieldset className="flex flex-col gap-6 border-t pt-6">
            <div className="flex flex-col gap-1 mb-6">
              <Skeleton width="280px" height="24px" />
              <Skeleton width="200px" height="16px" />
            </div>
            
            {/* Input fields in grid layout */}
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

          {/* Second input group */}
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
            
            {/* Checkbox items */}
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

          {/* Update options section */}
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
          onSubmit={handleFormDataSubmit}
          onSuccess={handleSuccess}
          onCancel={() => {
            removeModal(props.modalId)
          }}
        />
      )}
    </Modal>
  )
}

export const EditInputsButton = ({ ...props }: IEditInputs & IButtonAsButton) => {
  const { install } = useInstall()
  const { addModal } = useSurfaces()

  const showConfirmModal = () => {
    const confirmModal = (
      <ConfirmUpdateModal
        onConfirm={() => {
          // Show the main edit form
          const editModal = <EditInputsFormModal />
          addModal(editModal)
        }}
        onCancel={() => {
          // Just close the confirm modal, don't proceed
        }}
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