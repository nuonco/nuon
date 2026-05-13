import { useMemo, useRef } from 'react'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { CreateInstallForm } from '@/components/installs/forms/CreateInstallForm'
import type { TApp, TAppConfig } from '@/types'
import { FormSkeleton } from './FormSkeleton'

interface CreateInstallFromAppProps {
  app: TApp
  config: TAppConfig | undefined
  isLoading: boolean
  error: any
  submitError?: any
  isSubmitting: boolean
  onSelectApp: (app: TApp | undefined) => void
  onClose: () => void
  onSubmit: (formData: FormData) => Promise<any>
  formRef?: React.RefObject<HTMLFormElement>
  onRegisterClearDraft?: (clearFn: () => void) => void
}

export const CreateInstallFromApp = ({
  app,
  config,
  isLoading,
  error,
  submitError,
  isSubmitting,
  onSelectApp,
  onClose,
  onSubmit,
  formRef: externalFormRef,
  onRegisterClearDraft,
}: CreateInstallFromAppProps) => {
  const internalFormRef = useRef<HTMLFormElement>(null)
  const formRef = externalFormRef || internalFormRef

  const isDuplicateName = submitError?.error?.includes('duplicated key not allowed')
  const submitErrorMessage = useMemo(() => {
    if (!submitError) return undefined
    if (isDuplicateName) return 'Duplicate install names are not allowed. Please choose a different name.'
    return submitError.error || submitError.description || 'Unable to create install.'
  }, [submitError, isDuplicateName])

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

  if (isLoading) {
    return (
      <div>
        <div className="pb-4">
          <Button
            className="cursor-pointer w-fit text-primary-600 dark:text-primary-500 hover:text-primary-800 hover:dark:text-primary-400 focus:text-primary-800 focus:dark:text-primary-400 active:text-primary-900 active:dark:text-primary-600 focus-visible:rounded !bg-transparent !border-none !p-0 !h-auto font-medium"
            onClick={() => onSelectApp(undefined)}
          >
            <Icon variant="CaretLeftIcon" weight="bold" />
            Back
          </Button>
        </div>
        <FormSkeleton />
      </div>
    )
  }

  if (error || !config) {
    return (
      <div>
        <div className="pb-4">
          <Button
            className="cursor-pointer w-fit text-primary-600 dark:text-primary-500 hover:text-primary-800 hover:dark:text-primary-400 focus:text-primary-800 focus:dark:text-primary-400 active:text-primary-900 active:dark:text-primary-600 focus-visible:rounded !bg-transparent !border-none !p-0 !h-auto font-medium"
            onClick={() => onSelectApp(undefined)}
          >
            <Icon variant="CaretLeftIcon" weight="bold" />
            Back
          </Button>
        </div>
        <Banner theme="error">
          {error?.error || 'Unable to load app configuration'}
        </Banner>
      </div>
    )
  }

  return (
    <div>
      <div className="pb-4">
        <Button
          className="cursor-pointer w-fit text-primary-600 dark:text-primary-500 hover:text-primary-800 hover:dark:text-primary-400 focus:text-primary-800 focus:dark:text-primary-400 active:text-primary-900 active:dark:text-primary-600 focus-visible:rounded !bg-transparent !border-none !p-0 !h-auto font-medium"
          onClick={() => onSelectApp(undefined)}
        >
          <Icon variant="CaretLeftIcon" weight="bold" />
          Back
        </Button>
      </div>

      {submitErrorMessage && (
        <Banner theme="error" className="mb-6">
          {submitErrorMessage}
        </Banner>
      )}

      <CreateInstallForm
        ref={formRef}
        appId={app.id}
        platform={app.runner_config.app_runner_type as 'aws' | 'azure' | 'gcp'}
        nameError={isDuplicateName ? 'This name is already in use' : undefined}
        inputConfig={{
          ...config.input,
          input_groups: nestInputsUnderGroups(
            config.input?.input_groups,
            config.input?.inputs
          ),
        }}
        onSubmit={(formData: FormData) => onSubmit(formData)}
        onCancel={onClose}
        onRegisterClearDraft={onRegisterClearDraft}
      />
    </div>
  )
}
