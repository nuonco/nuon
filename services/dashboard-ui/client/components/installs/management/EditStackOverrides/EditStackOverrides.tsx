import { useForm, useStore } from '@tanstack/react-form'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { FormInput } from '@/components/common/form/FormInput'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'
import {
  editStackOverridesSchema,
  type CustomNestedStackEntry,
  type EditStackOverridesValues,
} from './schema'

interface IEditStackOverridesModal extends Omit<IModal, 'onSubmit'> {
  isPending: boolean
  error: TAPIError | null
  currentVpcUrl: string
  currentRunnerUrl: string
  currentCustomStacks: CustomNestedStackEntry[]
  appDefaultVpcUrl: string
  appDefaultRunnerUrl: string
  onSubmit: (data: {
    vpc_nested_template_url?: string
    runner_nested_template_url?: string
    custom_nested_stacks?: CustomNestedStackEntry[]
  }) => void
}

export const EditStackOverridesModal = ({
  isPending,
  error,
  currentVpcUrl,
  currentRunnerUrl,
  currentCustomStacks,
  appDefaultVpcUrl,
  appDefaultRunnerUrl,
  onSubmit,
  ...props
}: IEditStackOverridesModal) => {
  const form = useForm({
    defaultValues: {
      vpcUrl: currentVpcUrl,
      runnerUrl: currentRunnerUrl,
      customStacks: currentCustomStacks,
    } as EditStackOverridesValues,
    validators: {
      onMount: editStackOverridesSchema,
      onChange: editStackOverridesSchema,
    },
    onSubmit: ({ value }) => {
      const data: Parameters<typeof onSubmit>[0] = {}
      if (value.vpcUrl !== currentVpcUrl) {
        data.vpc_nested_template_url = value.vpcUrl
      }
      if (value.runnerUrl !== currentRunnerUrl) {
        data.runner_nested_template_url = value.runnerUrl
      }
      const validStacks = value.customStacks.filter(
        (s) => s.name && s.template_url
      )
      if (
        validStacks.length > 0 ||
        (currentCustomStacks.length > 0 && value.customStacks.length === 0)
      ) {
        data.custom_nested_stacks = validStacks
      }
      onSubmit(data)
    },
  })

  const canSubmit = useStore(form.store, (s) => s.canSubmit)

  return (
    <Modal
      size="lg"
      className="!max-h-[80vh]"
      childrenClassName="overflow-y-auto"
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon variant="StackSimpleIcon" size="24" />
          Edit stack overrides
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Saving overrides
          </span>
        ) : (
          'Save overrides'
        ),
        onClick: () => form.handleSubmit(),
        disabled: !canSubmit || isPending,
        variant: 'primary',
      }}
      {...props}
    >
      <form
        autoComplete="off"
        noValidate
        onSubmit={(e) => e.preventDefault()}
        className="flex flex-col gap-6"
      >
        <FormErrorBanner
          error={error}
          fallback="Unable to save stack overrides"
        />

        <Banner theme="info">
          <Text variant="body">
            Override the default stack template URLs for this install. Leave
            empty to use the app-level default.
          </Text>
        </Banner>

        <div className="flex flex-col gap-4">
          <form.Field name="vpcUrl">
            {(field) => (
              <FormInput
                field={field}
                id="vpc-template-url"
                type="text"
                disabled={isPending}
                labelProps={{ labelText: 'VPC nested template URL' }}
                placeholder={
                  appDefaultVpcUrl ||
                  'https://s3.amazonaws.com/bucket/vpc-template.yaml'
                }
                helperText={
                  appDefaultVpcUrl
                    ? `App default: ${appDefaultVpcUrl}`
                    : undefined
                }
              />
            )}
          </form.Field>

          <form.Field name="runnerUrl">
            {(field) => (
              <FormInput
                field={field}
                id="runner-template-url"
                type="text"
                disabled={isPending}
                labelProps={{ labelText: 'Runner nested template URL' }}
                placeholder={
                  appDefaultRunnerUrl ||
                  'https://s3.amazonaws.com/bucket/runner-template.yaml'
                }
                helperText={
                  appDefaultRunnerUrl
                    ? `App default: ${appDefaultRunnerUrl}`
                    : undefined
                }
              />
            )}
          </form.Field>
        </div>

        <div className="flex flex-col gap-3">
          <div className="flex items-center justify-between">
            <Text variant="subtext" weight="strong">
              Custom nested stacks
            </Text>
            <form.Field name="customStacks" mode="array">
              {(stacksField) => (
                <Button
                  variant="ghost"
                  size="sm"
                  type="button"
                  disabled={isPending}
                  onClick={() =>
                    stacksField.pushValue({
                      name: '',
                      template_url: '',
                      index: stacksField.state.value.length,
                    })
                  }
                >
                  <Icon variant="PlusIcon" />
                  Add stack
                </Button>
              )}
            </form.Field>
          </div>

          <form.Field name="customStacks" mode="array">
            {(stacksField) =>
              stacksField.state.value.length === 0 ? (
                <Text variant="subtext" theme="neutral">
                  No custom nested stack overrides configured.
                </Text>
              ) : (
                <div className="flex flex-col gap-3">
                  {stacksField.state.value.map((_, idx) => (
                    <div
                      key={idx}
                      className="flex flex-col gap-2 rounded-md border p-3"
                    >
                      <div className="flex items-center justify-between">
                        <Text variant="subtext" weight="strong">
                          Stack {idx + 1}
                        </Text>
                        <Button
                          variant="ghost"
                          size="sm"
                          type="button"
                          aria-label={`Remove stack ${idx + 1}`}
                          disabled={isPending}
                          onClick={() => stacksField.removeValue(idx)}
                        >
                          <Icon variant="TrashIcon" />
                        </Button>
                      </div>
                      <div className="grid grid-cols-[1fr_2fr_auto] gap-2 items-end">
                        <form.Field name={`customStacks[${idx}].name`}>
                          {(field) => (
                            <FormInput
                              field={field}
                              type="text"
                              disabled={isPending}
                              labelProps={{ labelText: 'Name' }}
                              placeholder="e.g. k8s_namespaces"
                            />
                          )}
                        </form.Field>
                        <form.Field name={`customStacks[${idx}].template_url`}>
                          {(field) => (
                            <FormInput
                              field={field}
                              type="text"
                              disabled={isPending}
                              labelProps={{ labelText: 'Template URL' }}
                              placeholder="https://s3.amazonaws.com/..."
                            />
                          )}
                        </form.Field>
                        <form.Field name={`customStacks[${idx}].index`}>
                          {(field) => (
                            <Input
                              type="number"
                              className="w-20"
                              value={String(field.state.value ?? 0)}
                              onChange={(e) =>
                                field.handleChange(
                                  parseInt(e.target.value) || 0
                                )
                              }
                              onBlur={field.handleBlur}
                              disabled={isPending}
                              labelProps={{ labelText: 'Index' }}
                            />
                          )}
                        </form.Field>
                      </div>
                    </div>
                  ))}
                </div>
              )
            }
          </form.Field>
        </div>
      </form>
    </Modal>
  )
}
