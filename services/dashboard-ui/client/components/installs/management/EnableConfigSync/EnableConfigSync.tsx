import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'

interface IEnableConfigSyncModal extends Omit<IModal, 'onSubmit'> {
  isManagedByConfig: boolean
  isPending: boolean
  error: any
  onSubmit: () => void
}

export const EnableConfigSyncModal = ({
  isManagedByConfig,
  isPending,
  error,
  onSubmit,
  ...props
}: IEnableConfigSyncModal) => {
  const buttonText = isManagedByConfig ? 'Disable Install Config Sync' : 'Enable Install Config Sync'
  const modalHeading = isManagedByConfig ? 'Disable Install Config Sync?' : 'Enable Install Config Sync?'

  return (
    <Modal
      heading={
        <Text
          flex
          className="gap-4"
          variant="h3"
          weight="strong"
        >
          <Icon variant="FileCloud" size="24" />
          {modalHeading}
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" />
            {isManagedByConfig ? 'Disabling...' : 'Enabling...'}
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant={isManagedByConfig ? 'ToggleRight' : 'ToggleLeft'} />
            {buttonText}
          </span>
        ),
        onClick: onSubmit,
        disabled: isPending,
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-3">
        {error ? (
          <Banner theme="error">
            {error?.error || 'An error happened, please refresh the page and try again.'}
          </Banner>
        ) : null}

        <Text variant="base">
          This Install can be managed via an Install Config file only after marking it as managed by Install Config.
        </Text>

        <Text variant="base">
          {isManagedByConfig
            ? 'Disabling this will stop any future syncs from the Install Config file.'
            : 'Enable this to allow syncing from an Install Config file.'}
        </Text>
      </div>
    </Modal>
  )
}
