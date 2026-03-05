import { type FormEvent, useRef, forwardRef, useEffect } from 'react'
import { RadioInput } from '@/components/common/form/RadioInput'
import { Text } from '@/components/common/Text'
import { useFormPersistence } from '@/hooks/use-form-persistence'
import { useSurfaces } from '@/hooks/use-surfaces'
import { InputConfigFields } from './shared/InputConfigFields'
import { ResumeDraftModal } from './shared/ResumeDraftModal'
import type { IUpdateInstallForm } from './shared/types'

const UpdateInstallOptions = () => {
  return (
    <fieldset className="flex flex-col gap-4 border-t pt-6">
      <legend className="flex flex-col gap-0 mb-4 pr-6">
        <span className="text-lg font-semibold">Update install resources</span>
        <span className="text-sm font-normal">
          Reprovision sandbox and redeploy components after updating install
          settings
        </span>
      </legend>

      <div className="flex gap-6 justify-start">
        <RadioInput
          name="form-control:update"
          value="skip"
          defaultChecked
          labelProps={{
            labelText: 'Skip updating resources',
            className:
              'hover:!bg-transparent focus:!bg-transparent active:!bg-transparent !px-0 !py-0 !w-fit',
          }}
        />
        <RadioInput
          name="form-control:update"
          value="update"
          labelProps={{
            labelText: 'Update all resources',
            className:
              'hover:!bg-transparent focus:!bg-transparent active:!bg-transparent !px-0 !py-0 !w-fit',
          }}
        />
      </div>
    </fieldset>
  )
}

export const UpdateInstallForm = forwardRef<
  HTMLFormElement,
  IUpdateInstallForm
>(
  (
    {
      install,
      platform,
      inputConfig,
      onSubmit,
      onSuccess,
      onCancel,
      onFormSubmit,
      onRegisterClearDraft,
    },
    ref
  ) => {
    const formRef = useRef<HTMLFormElement>(null)
    const draftShownRef = useRef(false)
    const { addModal, removeModal } = useSurfaces()

    const {
      hasDraft,
      draftTimestamp,
      draftValues,
      clearDraft,
      restoreDraft,
      formKey,
    } = useFormPersistence({
      storageKey: `install-update-draft:${install.id}`,
      formRef,
      enabled: true,
      configId: inputConfig?.id,
    })

    useEffect(() => {
      if (onRegisterClearDraft) {
        onRegisterClearDraft(clearDraft)
      }
    }, [onRegisterClearDraft, clearDraft])

    useEffect(() => {
      if (hasDraft && !draftShownRef.current && draftTimestamp) {
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
            onClose={() => {
              removeModal(modalId)
            }}
          />
        )
        modalId = addModal(modal)
      }
    }, [
      hasDraft,
      draftTimestamp,
      restoreDraft,
      clearDraft,
      addModal,
      removeModal,
    ])

    const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
      e.preventDefault()

      if (onFormSubmit) {
        onFormSubmit()
        return
      }

      const formData = new FormData(e.currentTarget)

      if (onSubmit) {
        try {
          const result = await onSubmit(formData)
          onSuccess?.(result)
        } catch (err) {
          console.error('Form submission error:', err)
        }
      }

      clearDraft()
    }

    return (
      <form
        key={formKey}
        ref={(node) => {
          formRef.current = node
          if (typeof ref === 'function') {
            ref(node)
          } else if (ref) {
            ref.current = node
          }
        }}
        onSubmit={handleSubmit}
        className="flex flex-col gap-8 justify-between focus:outline-none relative"
      >
        <div className="flex flex-col gap-8 max-w-4xl pb-12">
          <div className="flex flex-col gap-2">
            <Text variant="h3" weight="strong">
              Update {install.name}
            </Text>
            <Text variant="body" theme="neutral">
              Modify the configuration for this install.
            </Text>
          </div>

          {inputConfig && (
            <InputConfigFields
              key={formKey}
              inputConfig={inputConfig}
              install={install}
              draftValues={
                draftValues && Object.keys(draftValues).length > 0
                  ? draftValues
                  : undefined
              }
            />
          )}

          <UpdateInstallOptions />
        </div>
      </form>
    )
  }
)

UpdateInstallForm.displayName = 'UpdateInstallForm'
