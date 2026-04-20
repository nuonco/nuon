import { type FormEvent, forwardRef } from 'react'
import { Expand } from '@/components/common/Expand'
import { Input } from '@/components/common/form/Input'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { Text } from '@/components/common/Text'
import { InputConfigFields } from '../shared/InputConfigFields'
import { PlatformFields } from '../shared/PlatformFields'
import type { TAppInputConfig } from '@/types'

interface ICreateInstallFormPresentation {
  appId: string
  platform: 'aws' | 'azure' | 'gcp'
  nameError?: string
  inputConfig?: TAppInputConfig
  draftValues?: Record<string, string> | null
  formKey?: string
  clearDraft?: () => void
  onSubmit?: (formData: FormData) => Promise<any>
  onSuccess?: (result: any) => void
  onCancel: () => void
}

export const CreateInstallForm = forwardRef<
  HTMLFormElement,
  ICreateInstallFormPresentation
>(
  (
    {
      platform,
      nameError,
      inputConfig,
      draftValues,
      formKey,
      clearDraft,
      onSubmit,
      onSuccess,
    },
    ref
  ) => {
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
          clearDraft?.()
        } catch (err) {
          console.error('Form submission error:', err)
        }
      }
    }

    return (
      <form
        key={formKey}
        ref={ref}
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
              error={!!nameError}
              errorMessage={nameError}
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

          <Expand
            id="advanced-stack-overrides"
            heading="Advanced"
            headerClassName="!px-4 bg-code"
            className="mt-2 border rounded-md"
          >
            <div className="flex flex-col gap-6 p-4 border-t">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6 items-start">
                <span className="flex flex-col gap-0">
                  <Text variant="body" weight="strong">
                    VPC template URL override{' '}
                    <Text className="ml-1" variant="subtext" theme="neutral">
                      (optional)
                    </Text>
                  </Text>
                  <Text variant="subtext">
                    Override the app-level VPC nested CloudFormation template
                  </Text>
                </span>
                <Input
                  name="vpc_nested_template_url"
                  placeholder="https://s3.amazonaws.com/..."
                  defaultValue={draftValues?.vpc_nested_template_url || ''}
                />
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-6 items-start">
                <span className="flex flex-col gap-0">
                  <Text variant="body" weight="strong">
                    Runner template URL override{' '}
                    <Text className="ml-1" variant="subtext" theme="neutral">
                      (optional)
                    </Text>
                  </Text>
                  <Text variant="subtext">
                    Override the app-level runner nested CloudFormation template
                  </Text>
                </span>
                <Input
                  name="runner_nested_template_url"
                  placeholder="https://s3.amazonaws.com/..."
                  defaultValue={draftValues?.runner_nested_template_url || ''}
                />
              </div>
            </div>
          </Expand>

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
