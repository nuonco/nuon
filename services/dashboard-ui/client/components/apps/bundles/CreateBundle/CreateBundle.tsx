import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'

interface ICreateBundleModal extends Omit<IModal, 'onSubmit'> {
  appName?: string
  appConfigId?: string
  isPending: boolean
  error?: { error?: string } | null
  onSubmit: () => void
}

export const CreateBundleModal = ({
  appName,
  appConfigId,
  isPending,
  error,
  onSubmit,
  ...props
}: ICreateBundleModal) => (
  <Modal
    heading={
      <Text flex className="gap-4" variant="h3" weight="strong" theme="info">
        <Icon variant="PackageIcon" size="24" />
        Publish release
      </Text>
    }
    primaryActionTrigger={{
      children: isPending ? (
        <span className="flex items-center gap-2">
          <Icon variant="Loading" /> Publishing release
        </span>
      ) : (
        'Publish release'
      ),
      disabled: isPending || !appConfigId,
      onClick: onSubmit,
      variant: 'primary',
    }}
    {...props}
  >
    <div className="flex flex-col gap-3 mb-6">
      {error ? (
        <Banner theme="error">
          {error?.error || 'Unable to publish the release.'}
        </Banner>
      ) : null}
      <Text variant="base">
        This creates an immutable release from the latest config for {appName},
        then publishes a Linux AMD64 portable bundle. Publishing may take
        several minutes.
      </Text>
      {appConfigId ? (
        <Text variant="subtext" theme="neutral" flex className="gap-2">
          App config
          <Badge variant="code" size="sm">
            {appConfigId}
          </Badge>
        </Text>
      ) : (
        <Text variant="subtext" theme="neutral">
          No app config found. Sync your app config before publishing a release.
        </Text>
      )}
    </div>
  </Modal>
)
