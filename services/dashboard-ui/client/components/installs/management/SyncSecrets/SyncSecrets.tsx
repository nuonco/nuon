import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'

interface ISyncSecretsModal extends Omit<IModal, 'onSubmit'> {
  installName: string
  isPending: boolean
  error: any
  onSubmit: () => void
}

export const SyncSecretsModal = ({
  installName,
  isPending,
  error,
  onSubmit,
  ...props
}: ISyncSecretsModal) => {
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
          <Icon variant="KeyIcon" size="24" />
          Sync secrets?
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Syncing secrets
          </span>
        ) : (
          'Sync secrets'
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
          This will synchronize all secrets from your app configuration to the
          {installName} install environment.
        </Text>
      </div>
    </Modal>
  )
}
