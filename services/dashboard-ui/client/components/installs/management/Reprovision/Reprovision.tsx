import type { ReactNode } from 'react'
import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'

interface IReprovisionModal extends Omit<IModal, 'onSubmit'> {
  installName: string
  isPending: boolean
  error: any
  onSubmit: () => void
  roleSelector: ReactNode
}

export const ReprovisionModal = ({
  installName,
  isPending,
  error,
  onSubmit,
  roleSelector,
  ...props
}: IReprovisionModal) => {
  return (
    <Modal
      heading={
        <Text
          flex
          className="gap-4"
          variant="h3"
          weight="strong"
          theme="warn"
        >
          <Icon variant="ArrowURightUpIcon" size="24" />
          Reprovision install?
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Starting reprovision
          </span>
        ) : (
          'Reprovision install'
        ),
        onClick: onSubmit,
        variant: 'danger',
      }}
      {...props}
    >
      <div className="flex flex-col gap-4">
        {error ? (
          <Banner theme="error">
            {error?.error ||
              'Something went wrong. Try refreshing the page.'}
          </Banner>
        ) : null}
        <Text variant="base">
          Reprovisioning {installName} will recreate the stack and sandbox, and redeploy all components.
        </Text>
        <Banner theme="warn">
          <Text variant="body">
            <strong>Warning:</strong> Actions and deployments won't be
            available while the runner is recreated during the stack reprovision.
          </Text>
        </Banner>

        {roleSelector}
      </div>
    </Modal>
  )
}
