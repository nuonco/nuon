import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'

interface ISkipStepModal extends Omit<IModal, 'onSubmit'> {
  isPending: boolean
  error?: TAPIError | null
  onSubmit: () => void
}

export const SkipStepModal = ({
  isPending,
  error,
  onSubmit,
  ...props
}: ISkipStepModal) => {
  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="stronger">
          Skip step?
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Skipping step
          </span>
        ) : (
          'Skip step'
        ),
        onClick: onSubmit,
        disabled: isPending,
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-1">
        {error ? (
          <Banner theme="error">
            {error?.error ||
              'Something went wrong. Try refreshing the page.'}
          </Banner>
        ) : null}
        <Text variant="base">
          Skipping will bypass this step and continue the workflow with the
          remaining steps. Any actions or changes from this step will not be
          applied.
        </Text>
      </div>
    </Modal>
  )
}
