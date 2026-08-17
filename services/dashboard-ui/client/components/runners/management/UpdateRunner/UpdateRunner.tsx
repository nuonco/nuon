import { useForm, useStore } from '@tanstack/react-form'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { FormInput } from '@/components/common/form/FormInput'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'
import { updateRunnerSchema, type UpdateRunnerValues } from './schema'

interface IUpdateRunnerModal extends Omit<IModal, 'onSubmit'> {
  isPending: boolean
  error: TAPIError | null
  onSubmit: (tag: string) => void
  onClose: () => void
  modalHeading?: string
  inputLabel?: string
  inputPlaceholder?: string
  submitLabel?: string
}

export const UpdateRunnerModal = ({
  isPending,
  error,
  onSubmit,
  onClose,
  modalHeading = 'Update runner version',
  inputLabel = 'Enter the runner tag you\'d like to update to.',
  inputPlaceholder = 'runner tag',
  submitLabel = 'Update runner version',
  ...props
}: IUpdateRunnerModal) => {
  const form = useForm({
    defaultValues: { tag: '' } as UpdateRunnerValues,
    validators: { onMount: updateRunnerSchema, onChange: updateRunnerSchema },
    onSubmit: ({ value }) => onSubmit(value.tag),
  })

  const canSubmit = useStore(form.store, (s) => s.canSubmit)

  return (
    <Modal
      heading={
        <div className="flex flex-col gap-2">
          <Text
            flex
            className="gap-4"
            variant="h3"
            weight="strong"
          >
            <Icon variant="ArrowsCounterClockwiseIcon" size="24" />
            {modalHeading}
          </Text>
        </div>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" />
            Updating runner
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="ArrowsCounterClockwiseIcon" />
            {submitLabel}
          </span>
        ),
        disabled: !canSubmit || isPending,
        onClick: () => form.handleSubmit(),
        variant: 'primary' as const,
      }}
      onClose={onClose}
      {...props}
    >
      <form
        autoComplete="off"
        noValidate
        onSubmit={(e) => e.preventDefault()}
        className="flex flex-col gap-6"
      >
        <FormErrorBanner error={error} fallback="Unable to update runner" />

        <form.Field name="tag">
          {(field) => (
            <FormInput
              field={field}
              id="runner-tag"
              type="text"
              placeholder={inputPlaceholder}
              disabled={isPending}
              labelProps={{ labelText: inputLabel }}
            />
          )}
        </form.Field>
      </form>
    </Modal>
  )
}

interface IUpdateRunnerButton extends IButtonAsButton {
  onOpenModal: () => void
  label?: string
}

export const UpdateRunnerButton = ({
  onOpenModal,
  label = 'Update runner version',
  ...props
}: IUpdateRunnerButton) => {
  return (
    <Button
      onClick={() => onOpenModal()}
      {...props}
    >
      {props?.isMenuButton ? null : <Icon variant="ArrowsCounterClockwiseIcon" />}
      {label}
      {props?.isMenuButton ? <Icon variant="ArrowsCounterClockwiseIcon" /> : null}
    </Button>
  )
}
