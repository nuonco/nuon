import { useForm, useStore } from '@tanstack/react-form'
import { FormCheckbox } from '@/components/common/form/FormCheckbox'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { FormRadioGroup } from '@/components/common/form/FormRadioGroup'
import { FormSelect } from '@/components/common/form/FormSelect'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type {
  TAPIError,
  TAppBranchConfig,
  TAppBranchPreviewConfig,
  TInstall,
} from '@/types'
import {
  previewDefaultsFromConfig,
  previewDefaultsToConfig,
} from '@/components/branches/shared/PreviewDefaultsEditor'
import {
  previewConfigSchema,
  type PreviewConfigFormValues,
} from './schema'

export interface IPreviewConfigEditorModal extends Omit<IModal, 'onSubmit'> {
  currentConfig?: TAppBranchConfig
  installs: TInstall[]
  hasGithubVCS: boolean
  isPending?: boolean
  isLoading?: boolean
  error?: TAPIError | null
  onSubmit: (config: TAppBranchPreviewConfig) => void
  onCancel: () => void
}

export const PreviewConfigEditorModal = ({
  currentConfig,
  installs,
  hasGithubVCS,
  isPending = false,
  isLoading = false,
  error,
  onSubmit,
  onCancel,
  ...props
}: IPreviewConfigEditorModal) => {
  const initialDefaults = previewDefaultsFromConfig(
    currentConfig?.preview_config,
    installs
  )
  const initialValues: PreviewConfigFormValues = {
    mode: initialDefaults.mode,
    installId: initialDefaults.installId,
    setStatuses: initialDefaults.setStatuses,
    comment: initialDefaults.comment,
  }

  const form = useForm({
    defaultValues: initialValues,
    validators: {
      onMount: previewConfigSchema,
      onChange: previewConfigSchema,
    },
    onSubmit: ({ value }) =>
      onSubmit(
        previewDefaultsToConfig(
          {
            ...initialDefaults,
            ...value,
            installTargetMode: value.installId
              ? 'install'
              : initialDefaults.installTargetMode,
          },
          installs
        )
      ),
  })

  const canSubmit = useStore(form.store, (state) => state.canSubmit)
  const values = useStore(form.store, (state) => state.values)
  const isUnchanged =
    values.mode === initialValues.mode &&
    values.installId === initialValues.installId &&
    values.setStatuses === initialValues.setStatuses &&
    values.comment === initialValues.comment
  const installOptions = installs.map((install) => ({
    value: install.id,
    label: install.name,
  }))

  return (
    <Modal
      heading="Edit preview settings"
      primaryActionTrigger={{
        children: isPending ? 'Saving...' : 'Save changes',
        disabled: !canSubmit || isPending || isLoading || isUnchanged,
        onClick: () => form.handleSubmit(),
        tooltipProps: isUnchanged
          ? { tipContent: 'Change a setting before saving' }
          : isLoading
            ? { tipContent: 'Cannot save — installs are still loading' }
            : undefined,
        variant: 'primary',
      }}
      secondaryActionTrigger={{
        children: 'Cancel',
        disabled: isPending,
        onClick: onCancel,
      }}
      {...props}
    >
      <form
        autoComplete="off"
        noValidate
        onSubmit={(event) => event.preventDefault()}
        className="flex flex-col gap-6"
      >
        <FormErrorBanner
          error={error}
          fallback="Unable to save preview settings"
        />

        <form.Field name="mode">
          {(field) => (
            <FormRadioGroup
              field={field}
              label="Mode"
              disabled={isPending || isLoading}
              options={[
                { value: 'plan-only', label: 'Plan only' },
                { value: 'apply', label: 'Apply' },
                { value: 'build-only', label: 'Build only' },
              ]}
            />
          )}
        </form.Field>

        {values.mode !== 'build-only' ? (
          <form.Field name="installId">
            {(field) => (
              <FormSelect
                field={field}
                id="preview-default-install"
                options={installOptions}
                placeholder={isLoading ? 'Loading installs...' : 'Select an install'}
                disabled={isPending || isLoading || installOptions.length === 0}
                menuPlacement="bottom"
                labelProps={{ labelText: 'Default install' }}
                helperText={
                  initialDefaults.installTargetMode === 'labels' && !values.installId
                    ? 'Select an install to replace the current label selector.'
                    : 'Used for plan-only and apply preview runs.'
                }
              />
            )}
          </form.Field>
        ) : null}

        {hasGithubVCS ? (
          <div className="flex flex-col gap-3">
            <form.Field name="setStatuses">
              {(field) => (
                <FormCheckbox
                  field={field}
                  disabled={isPending || isLoading}
                  labelProps={{
                    labelText: (
                      <>
                        <Text weight="strong">Set commit statuses</Text>
                        <Text variant="subtext" theme="neutral">
                          Report preview progress and results to GitHub.
                        </Text>
                      </>
                    ),
                    labelTextProps: {
                      as: 'div',
                      className: 'flex flex-col gap-1',
                    },
                  }}
                  className="items-start"
                />
              )}
            </form.Field>
            <form.Field name="comment">
              {(field) => (
                <FormCheckbox
                  field={field}
                  disabled={isPending || isLoading}
                  labelProps={{
                    labelText: (
                      <>
                        <Text weight="strong">Comment on pull request</Text>
                        <Text variant="subtext" theme="neutral">
                          Post preview results to the pull request.
                        </Text>
                      </>
                    ),
                    labelTextProps: {
                      as: 'div',
                      className: 'flex flex-col gap-1',
                    },
                  }}
                  className="items-start"
                />
              )}
            </form.Field>
          </div>
        ) : null}
      </form>
    </Modal>
  )
}
