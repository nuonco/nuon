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
            {error?.error || 'Something went wrong. Try refreshing the page.'}
          </Banner>
        ) : null}
        <Text variant="base">
          Skipping bypasses this step and continues the workflow with the
          remaining steps. If a later step in this step&apos;s group applies
          changes regardless of this one, it can still run — check the
          workflow afterward.
        </Text>
      </div>
    </Modal>
  )
}
