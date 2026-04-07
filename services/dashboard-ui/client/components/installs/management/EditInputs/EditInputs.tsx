import { useRef, useState, type ReactNode } from 'react'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { Icon } from '@/components/common/Icon'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { UpdateInstallForm } from '@/components/installs/forms/UpdateInstallForm'
import type { TAppConfig, TInstall } from '@/types'

interface IConfirmUpdateModal extends IModal {
  isInstallManagedByConfig: boolean
  onConfirm: () => void
  onCancel: () => void
}

export const ConfirmUpdateModal = ({
  isInstallManagedByConfig,
  onConfirm,
  onCancel,
  ...props
}: IConfirmUpdateModal) => {
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
        onClick: onConfirm,
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
            onClick={onCancel}
            variant="ghost"
          >
            Cancel
          </Button>
        </div>
      </div>
    </Modal>
  )
}

interface IEditInputsFormModal extends IModal {
  install: TInstall
  config: TAppConfig | undefined
  isLoading: boolean
  error: any
  isSubmitting: boolean
  actionError: any
  onFormSubmit: () => void
  onClose: () => void
  formRef: React.RefObject<HTMLFormElement | null>
  clearDraftRef: React.MutableRefObject<(() => void) | null>
  selectedRole: string
  onRoleChange: (role: string) => void
  deployDependents: boolean
  onDeployDependentsChange: (checked: boolean) => void
  onMutate: (formData: FormData) => Promise<any>
}

export const EditInputsFormModal = ({
  install,
  config,
  isLoading,
  error,
  isSubmitting,
  actionError,
  onFormSubmit,
  onClose,
  formRef,
  clearDraftRef,
  selectedRole,
  onRoleChange,
  deployDependents,
  onDeployDependentsChange,
  onMutate,
  ...props
}: IEditInputsFormModal) => {
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

  return (
    <Modal
      {...props}
      size="lg"
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
              onClick: onFormSubmit,
              variant: 'primary',
            }
          : undefined
      }
      footerActions={
        <div className="flex flex-col gap-1 pl-4">
          <CheckboxInput
            checked={deployDependents}
            onChange={(e) => onDeployDependentsChange(e.target.checked)}
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
      onClose={onClose}
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
          onSubmit={(formData) => onMutate(formData)}
          onCancel={onClose}
          onRegisterClearDraft={(fn) => {
            clearDraftRef.current = fn
          }}
          selectedRole={selectedRole}
          onRoleChange={onRoleChange}
        />
      )}
    </Modal>
  )
}
