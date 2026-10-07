import { useEffect, useRef } from 'react'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import {
  InstallForm,
  useInstallForm,
  type InstallFormApi,
  type InstallFormValues,
} from '@/components/installs/forms/InstallForm'
import { ResumeDraftModal } from '@/components/installs/forms/shared/ResumeDraftModal'
import { useSurfaces } from '@/hooks/use-surfaces'
import type {
  TApp,
  TAppInputConfig,
  TAPIError,
  TCloudConnectionSummary,
} from '@/types'

export interface ICreateFormTriggerState {
  canSubmit: boolean
  submit: () => void
  form?: InstallFormApi
}

interface ICreateInstallFormFields {
  app: TApp
  inputConfig: TAppInputConfig
  cloudConnections?: TCloudConnectionSummary[]
  requireTargetAccount?: boolean
  defaultAutoApprove?: boolean
  defaultStackOnly?: boolean
  autoApproveDescription?: string
  submitError?: TAPIError | null
  validateName?: (name: string) => Promise<string | undefined>
  onSubmit: (values: InstallFormValues) => Promise<unknown> | false | void
  onStateChange: (state: ICreateFormTriggerState) => void
}

export const CreateInstallFormFields = ({
  app,
  inputConfig,
  cloudConnections,
  requireTargetAccount,
  defaultAutoApprove,
  defaultStackOnly,
  autoApproveDescription,
  submitError,
  validateName,
  onSubmit,
  onStateChange,
}: ICreateInstallFormFields) => {
  const { addModal, removeModal } = useSurfaces()
  const draftShownRef = useRef(false)
  const platform = app.runner_config?.app_runner_type as
    | 'aws'
    | 'azure'
    | 'gcp'
    | undefined

  const {
    form,
    canSubmit,
    isValidating,
    hasDraft,
    draftTimestamp,
    clearDraft,
    restoreDraft,
  } = useInstallForm({
    mode: 'create',
    platform,
    inputConfig,
    requireTargetAccount,
    defaultAutoApprove,
    defaultStackOnly,
    storageKey: `install-draft:${app.id}`,
    onSubmit: async (values) => {
      try {
        const result = await onSubmit(values)
        if (result !== false) clearDraft()
      } catch {
        return
      }
    },
  })

  useEffect(() => {
    onStateChange({
      canSubmit: canSubmit && !isValidating,
      submit: () => form.handleSubmit(),
      form,
    })
  }, [canSubmit, isValidating, form, onStateChange])

  useEffect(() => {
    if (!hasDraft || draftShownRef.current || !draftTimestamp) return
    draftShownRef.current = true

    let modalId: string
    const modal = (
      <ResumeDraftModal
        draftTimestamp={draftTimestamp}
        onResume={() => {
          restoreDraft()
          removeModal(modalId)
        }}
        onStartFresh={() => {
          clearDraft()
          draftShownRef.current = false
          removeModal(modalId)
        }}
        onClose={() => removeModal(modalId)}
      />
    )
    modalId = addModal(modal)
  }, [
    hasDraft,
    draftTimestamp,
    restoreDraft,
    clearDraft,
    addModal,
    removeModal,
  ])

  return (
    <div className="flex flex-col gap-6">
      <FormErrorBanner
        error={submitError}
        fallback="Unable to create install"
      />
      <InstallForm
        form={form}
        mode="create"
        platform={platform}
        inputConfig={inputConfig}
        cloudConnections={cloudConnections}
        requireTargetAccount={requireTargetAccount}
        autoApproveDescription={autoApproveDescription}
        validateName={validateName}
      />
    </div>
  )
}
