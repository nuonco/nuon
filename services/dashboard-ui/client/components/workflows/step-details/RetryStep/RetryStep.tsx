import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'

interface IRetryStepModal extends Omit<IModal, 'onSubmit'> {
  isPending: boolean
  error?: TAPIError | null
  onSubmit: () => void
}

export const RetryStepModal = ({
  isPending,
  error,
  onSubmit,
  ...props
}: IRetryStepModal) => {
  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="stronger">
          Retry step?
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Retrying step
          </span>
        ) : (
          'Retry step'
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
          Retrying will rerun this workflow step. If it succeeds, the workflow
          will continue from this point.
        </Text>
      </div>
    </Modal>
  )
}
