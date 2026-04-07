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
          theme="info"
        >
          <Icon variant="ArrowURightUp" size="24" />
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
        <div className="flex flex-col gap-1">
          <Text variant="base" weight="strong">
            Are you sure you want to reprovision {}?
          </Text>
          <Text variant="base">
            Reprovisioning will recreate all resources and deploy all components
            again.
          </Text>
        </div>

        {roleSelector}
      </div>
    </Modal>
  )
}
