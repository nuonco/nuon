import { useEffect, useRef } from 'react'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import {
  InstallForm,
  useInstallForm,
  type InstallFormValues,
} from '@/components/installs/forms/InstallForm'
import { ResumeDraftModal } from '@/components/installs/forms/shared/ResumeDraftModal'
import { useSurfaces } from '@/hooks/use-surfaces'
import type {
  TApp,
  TAppInputConfig,
  TAPIError,
  TAWSAccountConnection,
} from '@/types'

export interface ICreateFormTriggerState {
  canSubmit: boolean
  submit: () => void
}

interface ICreateInstallFormFields {
  app: TApp
  inputConfig: TAppInputConfig
  awsAccountConnections?: TAWSAccountConnection[]
  requireTargetAccount?: boolean
  defaultAutoApprove?: boolean
  autoApproveDescription?: string
  submitError?: TAPIError | null
  onSubmit: (values: InstallFormValues) => Promise<unknown> | void
  onStateChange: (state: ICreateFormTriggerState) => void
}

export const CreateInstallFormFields = ({
  app,
  inputConfig,
  awsAccountConnections,
  requireTargetAccount,
  defaultAutoApprove,
  autoApproveDescription,
  submitError,
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
    storageKey: `install-draft:${app.id}`,
    onSubmit: async (values) => {
      try {
        await onSubmit(values)
        clearDraft()
      } catch {
        // error surfaced via submitError → FormErrorBanner
      }
    },
  })

  useEffect(() => {
    onStateChange({ canSubmit, submit: () => form.handleSubmit() })
  }, [canSubmit, form, onStateChange])

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
        awsAccountConnections={awsAccountConnections}
        requireTargetAccount={requireTargetAccount}
        autoApproveDescription={autoApproveDescription}
      />
    </div>
  )
}
