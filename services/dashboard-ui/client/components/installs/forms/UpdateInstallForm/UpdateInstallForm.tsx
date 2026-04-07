import { type FormEvent, useRef, forwardRef } from 'react'
import { RadioInput } from '@/components/common/form/RadioInput'
import { Text } from '@/components/common/Text'
import { InputConfigFields } from '../shared/InputConfigFields'
import { RoleSelector } from '@/components/roles/RoleSelector'
import type { IUpdateInstallForm } from '../shared/types'

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

interface IUpdateInstallFormPresentation extends IUpdateInstallForm {
  draftValues?: Record<string, string> | null
  formKey?: string
  clearDraft?: () => void
}

export const UpdateInstallForm = forwardRef<
  HTMLFormElement,
  IUpdateInstallFormPresentation
>(
  (
    {
      install,
      inputConfig,
      onSubmit,
      onSuccess,
      onFormSubmit,
      selectedRole,
      onRoleChange,
      draftValues,
      formKey,
      clearDraft,
    },
    ref
  ) => {
    const formRef = useRef<HTMLFormElement>(null)

    const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
      e.preventDefault()

      if (onFormSubmit) {
        onFormSubmit()
        return
      }

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
          clearDraft?.()
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
          <RoleSelector
            installId={install?.id}
            value={selectedRole || ''}
            onChange={onRoleChange || (() => {})}
            name="role"
          />
        </div>
      </form>
    )
  }
)

UpdateInstallForm.displayName = 'UpdateInstallForm'
