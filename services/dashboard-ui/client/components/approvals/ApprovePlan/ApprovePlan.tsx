import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'

interface IApprovePlanModal extends Omit<IModal, 'onSubmit'> {
  modalCopy: { title: string; heading: string; message: string }
  isPending: boolean
  error?: TAPIError | null
  onSubmit: () => void
}

export const ApprovePlanModal = ({
  modalCopy,
  isPending,
  error,
  onSubmit,
  ...props
}: IApprovePlanModal) => {
  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="stronger">
          {modalCopy.title}
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Approving plan
          </span>
        ) : (
          'Approve plan'
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
              'An error happened, please refresh the page and try again.'}
          </Banner>
        ) : null}
        <Text variant="base" weight="stronger">
          {modalCopy.heading}
        </Text>
        <Text variant="base">{modalCopy.message}</Text>
      </div>
    </Modal>
  )
}
