import { useCallback, useEffect, useMemo, useState } from 'react'
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
  defaultStackOnly?: boolean
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
  defaultStackOnly,
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
    () =>
      buildInstallDefaults({
        mode,
        inputConfig,
        install,
        defaultAutoApprove,
        defaultStackOnly,
      }),
    [mode, inputConfig, install, defaultAutoApprove, defaultStackOnly]
  )

  const validator = schema as unknown as FormValidateOrFn<InstallFormValues>
  const [restoredValues, setRestoredValues] = useState<InstallFormValues>()

  const form = useForm({
    defaultValues: restoredValues ?? defaults,
    validators: { onMount: validator, onChange: validator },
    onSubmit: ({ value }) => onSubmit(value),
  })

  const canSubmit = useStore(form.store, (s) => s.canSubmit)
  const isSubmitting = useStore(form.store, (s) => s.isSubmitting)
  const isValidating = useStore(form.store, (s) => s.isFieldsValidating)
  const values = useStore(form.store, (s) => s.values)

  const { hasDraft, draftTimestamp, draftValues, clearDraft } =
    useDraftPersistence<InstallFormValues>({
      storageKey: storageKey ?? '',
      values,
      enabled: !!storageKey,
      configId: inputConfig?.id,
    })

  const restoreDraft = useCallback(() => {
    setRestoredValues(mergeDraftValues(defaults, draftValues))
  }, [defaults, draftValues])

  useEffect(() => {
    if (!restoredValues) return
    void form.validate('mount')
  }, [form, restoredValues])

  return {
    form,
    canSubmit,
    isSubmitting,
    isValidating,
    hasDraft,
    draftTimestamp,
    clearDraft,
    restoreDraft,
  }
}

export type InstallFormApi = ReturnType<typeof useInstallForm>['form']
