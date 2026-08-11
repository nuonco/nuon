import { useForm, useStore } from '@tanstack/react-form'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { FormInput } from '@/components/common/form/FormInput'
import { FormTextarea } from '@/components/common/form/FormTextarea'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { ICreateNotebookBody } from '@/lib'
import type { TAPIError } from '@/types'
import { createNotebookSchema, type CreateNotebookValues } from './schema'

interface ICreateNotebookModal extends Omit<IModal, 'onSubmit'> {
  isPending: boolean
  error: TAPIError | null
  onSubmit: (body: ICreateNotebookBody) => void
}

export const CreateNotebookModal = ({
  isPending,
  error,
  onSubmit,
  ...props
}: ICreateNotebookModal) => {
  const form = useForm({
    defaultValues: { name: '', description: '' } as CreateNotebookValues,
    validators: { onMount: createNotebookSchema, onChange: createNotebookSchema },
    onSubmit: ({ value }) =>
      onSubmit({ name: value.name, description: value.description || undefined }),
  })

  const canSubmit = useStore(form.store, (s) => s.canSubmit)

  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon variant="NotebookIcon" size="24" />
          Create notebook
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Creating notebook
          </span>
        ) : (
          'Create notebook'
        ),
        disabled: !canSubmit || isPending,
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
        <FormErrorBanner error={error} fallback="Unable to create the notebook" />

        <form.Field name="name">
          {(field) => (
            <FormInput
              field={field}
              id="notebook-name"
              type="text"
              placeholder="e.g. Debug pods"
              maxLength={255}
              disabled={isPending}
              labelProps={{ labelText: 'Name' }}
            />
          )}
        </form.Field>

        <form.Field name="description">
          {(field) => (
            <FormTextarea
              field={field}
              id="notebook-description"
              placeholder="What this notebook is for"
              maxLength={2000}
              rows={3}
              disabled={isPending}
              labelProps={{ labelText: 'Description (optional)' }}
            />
          )}
        </form.Field>
      </form>
    </Modal>
  )
}
