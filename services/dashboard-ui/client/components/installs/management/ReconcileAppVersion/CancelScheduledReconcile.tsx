import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError, TInstallScheduledReconcile } from '@/types'

export interface ICancelScheduledReconcileModal
  extends Omit<IModal, 'onSubmit'> {
  pending: TInstallScheduledReconcile
  isPending: boolean
  error?: TAPIError | null
  onSubmit: () => void
}

export const CancelScheduledReconcileModal = ({
  pending,
  isPending,
  error,
  onSubmit,
  ...props
}: ICancelScheduledReconcileModal) => (
  <Modal
    heading={
      <Text flex className="gap-4" variant="h3" weight="strong" theme="error">
        <Icon variant="WarningIcon" size="24" />
        Cancel scheduled reconcile?
      </Text>
    }
    primaryActionTrigger={{
      children: isPending ? (
        <span className="flex items-center gap-2">
          <Icon variant="Loading" /> Canceling
        </span>
      ) : (
        'Cancel schedule'
      ),
      disabled: isPending,
      onClick: onSubmit,
      variant: 'danger',
    }}
    {...props}
  >
    <div className="flex flex-col gap-4">
      {error ? (
        <Banner theme="error">
          {error?.error || 'Unable to cancel the scheduled reconcile.'}
        </Banner>
      ) : null}
      <Text variant="base">
        This cancels the reconcile scheduled for{' '}
        <Time time={pending.scheduled_at} format="long-datetime" />. The install
        will not be reconciled at that time.
      </Text>
    </div>
  </Modal>
)
