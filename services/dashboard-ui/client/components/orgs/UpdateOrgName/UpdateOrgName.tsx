import { useForm, useStore } from '@tanstack/react-form'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { FormInput } from '@/components/common/form/FormInput'
import type { TAPIError } from '@/types'
import { updateOrgNameSchema, type UpdateOrgNameValues } from './schema'

export const UpdateOrgNameForm = ({
  currentName,
  isPending,
  error,
  onSubmit,
}: {
  currentName: string
  isPending: boolean
  error: TAPIError | null
  onSubmit: (name: string) => void
}) => {
  const form = useForm({
    defaultValues: { name: currentName } as UpdateOrgNameValues,
    validators: {
      onMount: updateOrgNameSchema,
      onChange: updateOrgNameSchema,
    },
    onSubmit: ({ value }) => {
      const name = value.name.trim()
      if (!name || name === currentName || isPending) return
      onSubmit(name)
    },
  })

  const canSubmit = useStore(form.store, (state) => state.canSubmit)
  const name = useStore(form.store, (state) => state.values.name)
  const unchanged = name.trim() === currentName

  return (
    <form
      autoComplete="off"
      noValidate
      className="flex max-w-md flex-col gap-4"
      onSubmit={(event) => {
        event.preventDefault()
        void form.handleSubmit()
      }}
    >
      <FormErrorBanner
        error={error}
        fallback="Unable to update the organization name"
      />

      <form.Field name="name">
        {(field) => (
          <FormInput
            field={field}
            id="org-name"
            type="text"
            placeholder="acme"
            disabled={isPending}
            helperText="Must be unique across organizations."
            labelProps={{ labelText: 'Organization name' }}
          />
        )}
      </form.Field>

      <Button
        type="submit"
        variant="primary"
        disabled={!canSubmit || isPending || unchanged}
        tooltipProps={
          unchanged && !isPending
            ? { tipContent: 'No changes to save' }
            : undefined
        }
      >
        {isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Saving
          </span>
        ) : (
          'Save'
        )}
      </Button>
    </form>
  )
}
