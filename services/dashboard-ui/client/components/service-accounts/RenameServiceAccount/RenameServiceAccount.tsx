import { useForm, useStore } from '@tanstack/react-form'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { FormInput } from '@/components/common/form/FormInput'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'
import {
  renameServiceAccountSchema,
  type RenameServiceAccountValues,
} from './schema'

export const RenameServiceAccountModal = ({
  accountIdentity,
  currentName,
  isPending,
  error,
  onSubmit,
  ...props
}: {
  accountIdentity: string
  currentName: string
  isPending: boolean
  error: TAPIError | null
  onSubmit: (params: { name: string }) => void
} & Omit<IModal, 'onSubmit'>) => {
  const form = useForm({
    defaultValues: { name: currentName } as RenameServiceAccountValues,
    validators: {
      onMount: renameServiceAccountSchema,
      onChange: renameServiceAccountSchema,
    },
    onSubmit: ({ value }) => onSubmit({ name: value.name.trim() }),
  })

  const canSubmit = useStore(form.store, (s) => s.canSubmit)
  const name = useStore(form.store, (s) => s.values.name)

  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon variant="PencilSimpleIcon" size="24" />
          Rename service account
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Saving
          </span>
        ) : (
          'Save'
        ),
        disabled: !canSubmit || isPending || name === currentName,
        onClick: () => form.handleSubmit(),
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
          fallback="Unable to rename service account"
        />

        <Text>
          Rename <strong>{accountIdentity}</strong>.
        </Text>

        <form.Field name="name">
          {(field) => (
            <FormInput
              field={field}
              id="service-account-rename"
              type="text"
              placeholder="e.g. ci-deploy"
              disabled={isPending}
              labelProps={{ labelText: 'Name' }}
            />
          )}
        </form.Field>
      </form>
    </Modal>
  )
}
