import { useForm } from '@tanstack/react-form'
import { z } from 'zod'
import { Text } from '@/components/common/Text'
import { FormCheckbox } from '@/components/common/form/FormCheckbox'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'

const triggerBranchRunSchema = z.object({ force: z.boolean() })

export interface ITriggerBranchRunModal extends Omit<IModal, 'onSubmit'> {
  branchName: string
  planOnly: boolean
  isPending: boolean
  error?: TAPIError | null
  onConfirm: (force: boolean) => void
}

export const TriggerBranchRunModal = ({
  branchName,
  planOnly,
  isPending,
  error,
  onConfirm,
  ...props
}: ITriggerBranchRunModal) => {
  const form = useForm({
    defaultValues: { force: false },
    validators: { onSubmit: triggerBranchRunSchema },
    onSubmit: ({ value }) => {
      if (!isPending) onConfirm(value.force)
    },
  })

  return (
    <Modal
      heading={planOnly ? 'Trigger preview run' : 'Trigger run'}
      primaryActionTrigger={{
        children: isPending
          ? 'Triggering...'
          : planOnly
            ? 'Trigger preview'
            : 'Trigger run',
        disabled: isPending,
        onClick: () => form.handleSubmit(),
        variant: 'primary',
      }}
      {...props}
    >
      <Text>
        {planOnly
          ? `This starts a plan-only preview run for "${branchName}". Nothing will be deployed.`
          : `This starts a new run for "${branchName}" and deploys the current configuration to its install groups. This may take a few minutes.`}
      </Text>
      <form.Field name="force">
        {(field) => (
          <FormCheckbox
            field={field}
            disabled={isPending}
            className="mt-1"
            labelProps={{
              className: 'items-start',
              labelText: (
                <>
                  <Text as="p" weight="strong">
                    Rebuild all components and sandbox
                  </Text>
                  <Text
                    as="p"
                    variant="subtext"
                    theme="neutral"
                    className="mt-1"
                  >
                    Includes unchanged resources. Takes longer, and any build
                    failure stops the run.
                  </Text>
                </>
              ),
              labelTextProps: { as: 'div' },
            }}
          />
        )}
      </form.Field>
      <FormErrorBanner error={error} fallback="Unable to trigger run." />
    </Modal>
  )
}
