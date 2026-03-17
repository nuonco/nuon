import { type FormEvent, useRef, forwardRef, useEffect } from 'react'
import { Banner } from '@/components/common/Banner'
import { Input } from '@/components/common/form/Input'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { Text } from '@/components/common/Text'
import { useFormPersistence } from '@/hooks/use-form-persistence'
import { useSurfaces } from '@/hooks/use-surfaces'
import { InputConfigFields } from './shared/InputConfigFields'
import { PlatformFields } from './shared/PlatformFields'
import { ResumeDraftModal } from './shared/ResumeDraftModal'
import type { ICreateInstallForm } from './shared/types'

export const CreateInstallForm = forwardRef<
  HTMLFormElement,
  ICreateInstallForm
>(
  (
    {
      appId,
      platform,
      inputConfig,
      onSubmit,
      onSuccess,
      onCancel,
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
      storageKey: `install-draft:${appId}`,
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

      const form = e.currentTarget
      const firstInvalid = form.querySelector<HTMLElement>(':invalid:not(fieldset):not(form)')
      if (firstInvalid) {
        firstInvalid.scrollIntoView({ behavior: 'smooth', block: 'center' })
        firstInvalid.focus()
        form.reportValidity()
        return
      }

      const formData = new FormData(form)

      if (onSubmit) {
        try {
          const result = await onSubmit(formData)
          onSuccess?.(result)
          clearDraft()
        } catch (err) {
          console.error('Form submission error:', err)
        }
      }
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
        noValidate
        onSubmit={handleSubmit}
        className="flex flex-col min-h-[50vh] gap-8 justify-between focus:outline-none relative"
      >
        <div className="flex flex-col gap-8 max-w-4xl pb-12">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6 items-start">
            <span className="flex flex-col gap-0">
              <Text variant="body" weight="strong">
                Install name{' '}
                <Text className="ml-1" variant="subtext" theme="error">
                  {'*'}
                </Text>
              </Text>
              <Text variant="subtext" className="max-w-72">
                A unique name for this install
              </Text>
            </span>
            <Input
              id="install-name"
              name="name"
              placeholder="Enter install name"
              required
              defaultValue={draftValues?.name || ''}
            />
          </div>

          {platform && (
            <PlatformFields platform={platform} draftValues={draftValues} />
          )}

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6 items-start">
            <span className="flex flex-col gap-0">
              <Text variant="body" weight="strong">
                Deployment approval
              </Text>
              <Text variant="subtext" className="max-w-72">
                Choose how deployments should be approved
              </Text>
            </span>
            <CheckboxInput
              name="auto-approve"
              className="mt-[6px]"
              defaultChecked={draftValues?.['auto-approve'] === 'on' || false}
              labelProps={{
                className: 'items-start',
                labelText: (
                  <div className="flex flex-col gap-1">
                    <Text variant="body" weight="stronger">
                      Auto-approve changes
                    </Text>
                    <Text variant="subtext" theme="neutral">
                      Automatically approve and apply all future changes without
                      manual confirmation. You can change this later in the
                      install settings.
                    </Text>
                  </div>
                ),
              }}
            />
          </div>

          {inputConfig && (
            <InputConfigFields
              inputConfig={inputConfig}
              draftValues={draftValues}
            />
          )}
        </div>
      </form>
    )
  }
)

CreateInstallForm.displayName = 'CreateInstallForm'
