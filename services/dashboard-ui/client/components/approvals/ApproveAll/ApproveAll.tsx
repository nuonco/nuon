import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'
import { toSentenceCase } from '@/utils/string-utils'

interface IApproveAllModal extends Omit<IModal, 'onSubmit'> {
  pendingSteps: { id: string; name: string }[]
  isPending: boolean
  error?: TAPIError | null
  onSubmit: () => void
}

export const ApproveAllModal = ({
  pendingSteps,
  isPending,
  error,
  onSubmit,
  ...props
}: IApproveAllModal) => {
  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="stronger">
          Approve all plans?
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Approving changes
          </span>
        ) : (
          'Approve all'
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
          Are you sure you want to approve all proposed changes?
        </Text>
        <Text variant="base">
          Approving all plans will immediately apply every set of outlined
          changes across the workflow.
        </Text>
        <Text className="mt-3" variant="base" weight="stronger">
          Step to approve
        </Text>
        <div className="flex flex-wrap gap-2">
          {pendingSteps.map((s) => (
            <Badge variant="code" key={s.id} size="sm">
              {toSentenceCase(s.name)}
            </Badge>
          ))}
        </div>
      </div>
    </Modal>
  )
}
