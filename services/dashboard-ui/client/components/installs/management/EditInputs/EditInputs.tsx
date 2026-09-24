import { useEffect, useRef } from 'react'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { RadioInput } from '@/components/common/form/RadioInput'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { InstallForm } from '@/components/installs/forms/InstallForm'
import {
  buildInstallInputDefaults,
  useInstallForm,
  type InstallFormValues,
} from '@/components/installs/forms/InstallForm'
import { ResumeDraftModal } from '@/components/installs/forms/shared/ResumeDraftModal'
import { useSurfaces } from '@/hooks/use-surfaces'
import type { TAppInputConfig, TAPIError, TInstall } from '@/types'

export interface IEditInputsUpdatePayload {
  name?: string
  inputs: Record<string, string>
  role?: string
  deployDependents: boolean
  inputsOnly: boolean
}

interface IEditInstallModal extends Omit<IModal, 'onSubmit'> {
  install: TInstall
  inputConfig: TAppInputConfig
  showNameField?: boolean
  isSubmitting: boolean
  submitError: TAPIError | null
  onSubmitName: (name: string) => Promise<unknown>
  onSubmitInputs: (payload: IEditInputsUpdatePayload) => Promise<unknown>
}

const toInputsPayload = (
  inputs: InstallFormValues['inputs']
): Record<string, string> => {
  const out: Record<string, string> = {}
  for (const [key, value] of Object.entries(inputs)) {
    out[key] = typeof value === 'boolean' ? String(value) : value
  }
  return out
}

export const EditInstallModal = ({
  install,
  inputConfig,
  showNameField,
  isSubmitting,
  submitError,
  onSubmitName,
  onSubmitInputs,
  ...props
}: IEditInstallModal) => {
  const { addModal, removeModal } = useSurfaces()
  const draftShownRef = useRef(false)

  const { form, canSubmit, hasDraft, draftTimestamp, clearDraft, restoreDraft } =
    useInstallForm({
      mode: 'edit',
      inputConfig,
      install,
      showNameField,
      storageKey: `install-update-draft:${install.id}`,
      onSubmit: async (values) => {
        const nameChanged =
          !!showNameField && values.name.trim() !== (install.name ?? '')
        const baseline = buildInstallInputDefaults(inputConfig, install)
        const inputsChanged =
          JSON.stringify(values.inputs) !== JSON.stringify(baseline)
        const roleSet = !!values.role

        try {
          if (nameChanged && !inputsChanged && !roleSet) {
            await onSubmitName(values.name.trim())
          } else {
            await onSubmitInputs({
              name: nameChanged ? values.name.trim() : undefined,
              inputs: toInputsPayload(values.inputs),
              role: values.role || undefined,
              deployDependents: values.deployDependents,
              inputsOnly: values.inputsOnly,
            })
          }
          clearDraft()
        } catch {
          // error surfaced via submitError → FormErrorBanner
        }
      },
    })

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
    <Modal
      {...props}
      size="lg"
      className="!max-h-[80vh]"
      childrenClassName="overflow-y-auto"
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon variant="PencilSimpleLineIcon" size="24" />
          {showNameField ? 'Edit install' : 'Edit install inputs'}
        </Text>
      }
      primaryActionTrigger={{
        children: isSubmitting ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" />
            {showNameField ? 'Updating install' : 'Updating inputs'}
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="CubeIcon" />
            {showNameField ? 'Update install' : 'Update inputs'}
          </span>
        ),
        disabled: !canSubmit || isSubmitting,
        onClick: () => form.handleSubmit(),
        variant: 'primary',
      }}
      footerActions={
        <form.Subscribe
          selector={(state) => ({
            inputsOnly: state.values.inputsOnly,
            deployDependents: state.values.deployDependents,
          })}
        >
          {({ inputsOnly, deployDependents }) => {
            const mode = inputsOnly
              ? 'inputs-only'
              : deployDependents
                ? 'deploy-dependents'
                : 'deploy'
            const setMode = (next: typeof mode) => {
              form.setFieldValue('inputsOnly', next === 'inputs-only')
              form.setFieldValue(
                'deployDependents',
                next === 'deploy-dependents'
              )
            }
            const radioClass =
              'hover:!bg-transparent focus:!bg-transparent active:!bg-transparent !px-0 !py-1 gap-4 max-w-none items-start'

            return (
              <div
                className="flex flex-col gap-1 pl-4"
                role="radiogroup"
                aria-label="How to apply input changes"
              >
                {(
                  [
                    {
                      value: 'inputs-only',
                      title: 'Save inputs only',
                      description:
                        'Save the values without deploying components or reprovisioning the sandbox.',
                    },
                    {
                      value: 'deploy',
                      title: 'Deploy affected components',
                      description:
                        'Deploy components whose inputs changed, without dependents.',
                    },
                    {
                      value: 'deploy-dependents',
                      title: 'Deploy dependents',
                      description:
                        'Deploy all dependents as well as the affected components.',
                    },
                  ] as const
                ).map((option) => (
                  <RadioInput
                    key={option.value}
                    name="edit-inputs-mode"
                    value={option.value}
                    checked={mode === option.value}
                    onChange={() => setMode(option.value)}
                    labelProps={{
                      className: radioClass,
                      labelText: (
                        <span className="flex flex-col gap-1">
                          <Text
                            variant="base"
                            weight="stronger"
                            className="!leading-none"
                          >
                            {option.title}
                          </Text>
                          <Text
                            variant="subtext"
                            theme="neutral"
                            className="!leading-none"
                          >
                            {option.description}
                          </Text>
                        </span>
                      ),
                      labelTextProps: { as: 'div' },
                    }}
                  />
                ))}
              </div>
            )
          }}
        </form.Subscribe>
      }
    >
      <div className="flex flex-col gap-6">
        <FormErrorBanner error={submitError} fallback="Unable to update install" />
        <InstallForm
          form={form}
          mode="edit"
          inputConfig={inputConfig}
          install={install}
          installId={install.id}
          showNameField={showNameField}
        />
      </div>
    </Modal>
  )
}
