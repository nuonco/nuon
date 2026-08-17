import { useForm, useStore } from '@tanstack/react-form'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { FormInput } from '@/components/common/form/FormInput'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'
import { createAppSchema, type CreateAppValues } from './schema'

export const CreateAppModal = ({
  isSubmitting,
  error,
  onSubmit,
  ...props
}: {
  isSubmitting: boolean
  error: TAPIError | null
  onSubmit: (body: CreateAppValues) => void
} & Omit<IModal, 'onSubmit'>) => {
  const form = useForm({
    defaultValues: { name: '' } as CreateAppValues,
    validators: { onMount: createAppSchema, onChange: createAppSchema },
    onSubmit: ({ value }) => onSubmit(value),
  })

  const canSubmit = useStore(form.store, (s) => s.canSubmit)

  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon variant="AppWindowIcon" size="24" />
          Create app
        </Text>
      }
      primaryActionTrigger={{
        children: isSubmitting ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Creating app
          </span>
        ) : (
          'Create app'
        ),
        disabled: !canSubmit || isSubmitting,
        onClick: () => form.handleSubmit(),
        variant: 'primary',
      }}
      {...props}
    >
      <form
        autoComplete="off"
        noValidate
        onSubmit={(e) => e.preventDefault()}
        className="flex flex-col gap-4"
      >
        <FormErrorBanner error={error} fallback="Unable to create app" />

        <form.Field name="name">
          {(field) => (
            <FormInput
              field={field}
              id="app-name"
              type="text"
              placeholder="my-app"
              maxLength={255}
              disabled={isSubmitting}
              labelProps={{ labelText: 'App name' }}
            />
          )}
        </form.Field>
      </form>
    </Modal>
  )
}
