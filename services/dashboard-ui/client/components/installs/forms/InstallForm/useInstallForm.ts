import { useCallback, useMemo } from 'react'
import { useForm, useStore } from '@tanstack/react-form'
import type { FormValidateOrFn } from '@tanstack/form-core'
import { useDraftPersistence } from '@/hooks/use-draft-persistence'
import type { TAppInputConfig, TInstall } from '@/types'
import { buildInstallDefaults, mergeDraftValues } from './defaults'
import {
  buildInstallSchema,
  type InstallFormMode,
  type InstallFormValues,
  type InstallPlatform,
} from './schema'

export interface UseInstallFormParams {
  mode: InstallFormMode
  platform?: InstallPlatform
  inputConfig?: TAppInputConfig
  install?: TInstall
  requireTargetAccount?: boolean
  showNameField?: boolean
  defaultAutoApprove?: boolean
  storageKey?: string
  onSubmit: (values: InstallFormValues) => void | Promise<unknown>
}

export function useInstallForm({
  mode,
  platform,
  inputConfig,
  install,
  requireTargetAccount,
  showNameField,
  defaultAutoApprove,
  storageKey,
  onSubmit,
}: UseInstallFormParams) {
  const schema = useMemo(
    () =>
      buildInstallSchema({
        mode,
        platform,
        inputConfig,
        requireTargetAccount,
        showNameField,
      }),
    [mode, platform, inputConfig, requireTargetAccount, showNameField]
  )

  const defaults = useMemo(
    () => buildInstallDefaults({ mode, inputConfig, install, defaultAutoApprove }),
    [mode, inputConfig, install, defaultAutoApprove]
  )

  const validator = schema as unknown as FormValidateOrFn<InstallFormValues>

  const form = useForm({
    defaultValues: defaults,
    validators: { onMount: validator, onChange: validator },
    onSubmit: ({ value }) => onSubmit(value),
  })

  const canSubmit = useStore(form.store, (s) => s.canSubmit)
  const isSubmitting = useStore(form.store, (s) => s.isSubmitting)
  const values = useStore(form.store, (s) => s.values)

  const { hasDraft, draftTimestamp, draftValues, clearDraft } =
    useDraftPersistence<InstallFormValues>({
      storageKey: storageKey ?? '',
      values,
      enabled: !!storageKey,
      configId: inputConfig?.id,
    })

  const restoreDraft = useCallback(() => {
    form.reset(mergeDraftValues(defaults, draftValues))
  }, [form, defaults, draftValues])

  return {
    form,
    canSubmit,
    isSubmitting,
    hasDraft,
    draftTimestamp,
    clearDraft,
    restoreDraft,
  }
}

export type InstallFormApi = ReturnType<typeof useInstallForm>['form']
