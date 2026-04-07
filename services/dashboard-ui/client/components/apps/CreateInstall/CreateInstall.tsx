import { useRef, forwardRef } from 'react'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { CreateInstallForm } from '@/components/installs/forms/CreateInstallForm'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAppConfig } from '@/types'
import type { TAPIError } from '@/types'

interface ICreateInstall {}

export const FormSkeleton = () => {
  return (
    <div className="flex flex-col gap-8 max-w-4xl">
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6 items-start">
        <span className="flex flex-col gap-1">
          <Skeleton width="100px" height="16px" />
          <Skeleton width="160px" height="14px" />
        </span>
        <Skeleton width="100%" height="40px" />
      </div>

      <fieldset className="flex flex-col gap-6 border-t pt-6">
        <Skeleton width="140px" height="24px" />
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 items-start">
          <span className="flex flex-col gap-1">
            <Skeleton width="130px" height="16px" />
          </span>
          <Skeleton width="100%" height="40px" />
        </div>
      </fieldset>

      <fieldset className="flex flex-col gap-6 border-t pt-6">
        <div className="flex flex-col gap-1 mb-6">
          <Skeleton width="220px" height="24px" />
          <Skeleton width="280px" height="16px" />
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 items-start">
          <span className="flex flex-col gap-1">
            <Skeleton width="80px" height="16px" />
            <Skeleton width="200px" height="14px" />
          </span>
          <Skeleton width="100%" height="40px" />
        </div>
      </fieldset>

      <fieldset className="flex flex-col gap-6 border-t pt-6">
        <div className="flex flex-col gap-1 mb-6">
          <Skeleton width="180px" height="24px" />
          <Skeleton width="140px" height="16px" />
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 items-start">
          <span className="flex flex-col gap-1">
            <Skeleton width="90px" height="16px" />
            <Skeleton width="160px" height="14px" />
          </span>
          <Skeleton width="100%" height="40px" />
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 items-start">
          <span className="flex flex-col gap-1">
            <Skeleton width="110px" height="16px" />
            <Skeleton width="240px" height="14px" />
          </span>
          <Skeleton width="100%" height="40px" />
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 items-start">
          <span className="flex flex-col gap-1">
            <Skeleton width="85px" height="16px" />
            <Skeleton width="300px" height="14px" />
          </span>
          <Skeleton width="100%" height="40px" />
        </div>
      </fieldset>
    </div>
  )
}

interface ICreateInstallModal extends ICreateInstall, IModal {
  isLoading: boolean
  hasError: boolean
  configsError?: TAPIError | null
  configError?: TAPIError | null
  config?: TAppConfig
  configs?: TAppConfig[]
  isSubmitting: boolean
  onFormSubmit?: () => void
  appId: string
  platform: 'aws' | 'azure' | 'gcp'
  onSubmitAction: (formData: FormData) => Promise<any>
  onCancel: () => void
}

export const CreateInstallModal = ({
  isLoading,
  hasError,
  configsError,
  configError,
  config,
  configs,
  isSubmitting,
  onFormSubmit,
  appId,
  platform,
  onSubmitAction,
  onCancel,
  ...props
}: ICreateInstallModal) => {
  const formRef = useRef<HTMLFormElement>(null)

  const canSubmit = !isLoading && !hasError && config

  const handleFormSubmit = () => {
    if (formRef.current) {
      formRef.current.requestSubmit()
    }
  }

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
      heading={
        <Text
          flex
          className="gap-4"
          variant="h3"
          weight="strong"
        >
          <Icon variant="Cube" size="24" />
          Create install
        </Text>
      }
      size="lg"
      className="!max-h-[80vh]"
      childrenClassName="overflow-y-auto"
      primaryActionTrigger={
        canSubmit
          ? {
              children: isSubmitting ? (
                <span className="flex items-center gap-2">
                  <Icon variant="Loading" />
                  Creating install
                </span>
              ) : (
                <span className="flex items-center gap-2">
                  <Icon variant="Cube" />
                  Create install
                </span>
              ),
              disabled: isSubmitting,
              onClick: onFormSubmit || handleFormSubmit,
              variant: 'primary',
            }
          : undefined
      }
      {...props}
    >
      {isLoading ? (
        <FormSkeleton />
      ) : hasError ? (
        <Banner theme="error">
          {configsError?.error ||
            configError?.error ||
            'Unable to load app configuration'}
        </Banner>
      ) : (
        <CreateInstallForm
          ref={formRef}
          appId={appId}
          platform={platform}
          inputConfig={{
            ...config?.input,
            input_groups: nestInputsUnderGroups(
              config?.input?.input_groups,
              config?.input?.inputs
            ),
          }}
          onSubmit={(formData) => onSubmitAction(formData)}
          onCancel={onCancel}
        />
      )}
    </Modal>
  )
}

export const CreateInstallButton = ({
  onClick,
  ...props
}: { onClick: () => void } & Omit<IButtonAsButton, 'onClick'>) => {
  return (
    <Button
      onClick={onClick}
      {...props}
    >
      <Icon variant="Cube" />
      Create install
    </Button>
  )
}
