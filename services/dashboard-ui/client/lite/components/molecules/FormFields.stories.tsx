import { useState } from 'react'
import { useForm, useStore } from '@tanstack/react-form'
import { z } from 'zod'
import { ComponentDocs } from '../__stories__/ComponentDocs'
import { Button } from '../atoms/Button'
import { Card } from '../atoms/Card'
import { Text } from '../atoms/Text'
import { FormCheckbox } from './FormCheckbox'
import { FormErrorBanner } from './FormErrorBanner'
import { FormInput } from './FormInput'
import { FormRadioGroup } from './FormRadioGroup'
import { FormSelect } from './FormSelect'
import { FormSwitch } from './FormSwitch'
import { FormTextarea } from './FormTextarea'

export default { title: 'lite/molecules/Form fields' }

const REGIONS = [
  {
    value: 'us-east-1',
    label: 'US East (N. Virginia)',
    description: 'us-east-1',
  },
  { value: 'us-west-2', label: 'US West (Oregon)', description: 'us-west-2' },
  { value: 'eu-west-1', label: 'Europe (Ireland)', description: 'eu-west-1' },
]

export const Overview = () => (
  <ComponentDocs
    name="Form fields"
    tier="molecule"
    summary="Thin TanStack Form adapters for every lite field."
    use={[
      'Bind fields with form.Field and pass the field object.',
      'Let Zod own all validation.',
    ]}
    avoid={[
      'Do not pass native required.',
      'Do not hold a second copy of field state in React.',
    ]}
    rules={[
      'Errors appear after a field is touched.',
      'API errors use FormErrorBanner inside the form.',
    ]}
    props={[
      {
        name: 'field',
        type: 'AnyFieldApi',
        description: 'TanStack field instance.',
      },
    ]}
  />
)

const InputDemo = () => {
  const form = useForm({ defaultValues: { name: '' } })
  return (
    <form
      className="max-w-sm p-8"
      noValidate
      onSubmit={(event) => event.preventDefault()}
    >
      <form.Field
        name="name"
        validators={{ onBlur: z.string().min(3, 'Use at least 3 characters') }}
      >
        {(field) => (
          <FormInput
            field={field}
            label="Install name"
            description="Lowercase letters, numbers and hyphens."
            placeholder="acme-production"
          />
        )}
      </form.Field>
    </form>
  )
}

export const InputField = () => <InputDemo />

const TextareaDemo = () => {
  const form = useForm({ defaultValues: { description: '' } })
  return (
    <form
      className="max-w-sm p-8"
      noValidate
      onSubmit={(event) => event.preventDefault()}
    >
      <form.Field
        name="description"
        validators={{
          onBlur: z.string().min(10, 'Add at least 10 characters'),
        }}
      >
        {(field) => (
          <FormTextarea
            field={field}
            label="Description"
            optional
            autoResize
            minRows={2}
            placeholder="What this install is for"
          />
        )}
      </form.Field>
    </form>
  )
}

export const TextareaField = () => <TextareaDemo />

const SelectDemo = () => {
  const form = useForm({ defaultValues: { region: '' } })
  return (
    <form
      className="max-w-sm p-8"
      noValidate
      onSubmit={(event) => event.preventDefault()}
    >
      <form.Field
        name="region"
        validators={{ onBlur: z.string().min(1, 'Choose a region') }}
      >
        {(field) => (
          <FormSelect
            field={field}
            label="Region"
            options={REGIONS}
            searchable
          />
        )}
      </form.Field>
    </form>
  )
}

export const SelectField = () => <SelectDemo />

const CheckboxDemo = () => {
  const form = useForm({ defaultValues: { confirmed: false } })
  return (
    <form
      className="max-w-sm p-8"
      noValidate
      onSubmit={(event) => event.preventDefault()}
    >
      <form.Field
        name="confirmed"
        validators={{
          onBlur: z
            .boolean()
            .refine(Boolean, 'Confirm the plan before continuing'),
        }}
      >
        {(field) => (
          <FormCheckbox
            field={field}
            label="I reviewed the plan"
            description="Confirm that the proposed infrastructure changes are expected."
          />
        )}
      </form.Field>
    </form>
  )
}

export const CheckboxField = () => <CheckboxDemo />

const RadioDemo = () => {
  const form = useForm({ defaultValues: { approval: '' } })
  return (
    <form
      className="max-w-sm p-8"
      noValidate
      onSubmit={(event) => event.preventDefault()}
    >
      <form.Field
        name="approval"
        validators={{ onBlur: z.string().min(1, 'Choose an approval mode') }}
      >
        {(field) => (
          <FormRadioGroup
            field={field}
            label="Approval mode"
            options={[
              {
                value: 'manual',
                label: 'Manual',
                description: 'Review every plan.',
              },
              {
                value: 'automatic',
                label: 'Automatic',
                description: 'Deploy matching plans.',
              },
              { value: 'policy', label: 'Policy based', disabled: true },
            ]}
          />
        )}
      </form.Field>
    </form>
  )
}

export const RadioGroupField = () => <RadioDemo />

const SwitchDemo = () => {
  const form = useForm({ defaultValues: { notifications: true } })
  return (
    <form
      className="max-w-sm p-8"
      noValidate
      onSubmit={(event) => event.preventDefault()}
    >
      <form.Field name="notifications">
        {(field) => (
          <FormSwitch
            field={field}
            label="Deploy notifications"
            description="Notify subscribed channels when deploys finish."
          />
        )}
      </form.Field>
    </form>
  )
}

export const SwitchField = () => <SwitchDemo />

const completeSchema = z.object({
  name: z.string().min(3, 'Use at least 3 characters'),
  description: z.string().max(120, 'Keep the description under 120 characters'),
  region: z.string().min(1, 'Choose a region'),
  approval: z.string().min(1, 'Choose an approval mode'),
  confirmed: z.boolean().refine(Boolean, 'Review and confirm the plan'),
  notifications: z.boolean(),
})

const CompleteDemo = () => {
  const [submitted, setSubmitted] = useState(false)
  const form = useForm({
    defaultValues: {
      name: '',
      description: '',
      region: '',
      approval: 'manual',
      confirmed: false,
      notifications: true,
    },
    validators: { onMount: completeSchema, onChange: completeSchema },
    onSubmit: () => setSubmitted(true),
  })
  const canSubmit = useStore(form.store, (state) => state.canSubmit)

  return (
    <Card className="m-8 max-w-lg">
      <form
        className="flex flex-col gap-5"
        autoComplete="off"
        noValidate
        onSubmit={(event) => {
          event.preventDefault()
          form.handleSubmit()
        }}
      >
        <span>
          <Text as="h2" variant="heading">
            Create install
          </Text>
          <Text variant="caption" color="tertiary">
            Configure the first deployment target.
          </Text>
        </span>
        <form.Field name="name">
          {(field) => (
            <FormInput
              field={field}
              label="Install name"
              placeholder="acme-production"
            />
          )}
        </form.Field>
        <form.Field name="description">
          {(field) => (
            <FormTextarea
              field={field}
              label="Description"
              optional
              autoResize
              minRows={2}
            />
          )}
        </form.Field>
        <form.Field name="region">
          {(field) => (
            <FormSelect
              field={field}
              label="Region"
              options={REGIONS}
              searchable
            />
          )}
        </form.Field>
        <form.Field name="approval">
          {(field) => (
            <FormRadioGroup
              field={field}
              label="Approval mode"
              options={[
                { value: 'manual', label: 'Manual' },
                { value: 'automatic', label: 'Automatic' },
              ]}
            />
          )}
        </form.Field>
        <form.Field name="notifications">
          {(field) => <FormSwitch field={field} label="Deploy notifications" />}
        </form.Field>
        <form.Field name="confirmed">
          {(field) => (
            <FormCheckbox field={field} label="I reviewed the configuration" />
          )}
        </form.Field>
        {submitted ? (
          <Text variant="caption" color="positive">
            Form submitted with valid values.
          </Text>
        ) : null}
        <div className="flex justify-end gap-2">
          <Button variant="ghost" onClick={() => form.reset()}>
            Reset
          </Button>
          <Button type="submit" variant="primary" disabled={!canSubmit}>
            Create install
          </Button>
        </div>
      </form>
    </Card>
  )
}

export const CompleteForm = () => <CompleteDemo />

export const ApiError = () => (
  <div className="max-w-lg p-8">
    <FormErrorBanner
      fallback="Unable to create the install"
      error={{
        error: 'Install creation failed',
        description: 'The selected region is not available for this account.',
        user_error: true,
      }}
    />
  </div>
)
