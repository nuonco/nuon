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
        Create bundle
      </Text>
    }
    primaryActionTrigger={{
      children: isPending ? (
        <span className="flex items-center gap-2">
          <Icon variant="Loading" /> Creating bundle
        </span>
      ) : (
        'Create bundle'
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
          {error?.error || 'Unable to create the bundle.'}
        </Banner>
      ) : null}
      <Text variant="base">
        This will package the latest config for {appName} into a single portable
        archive for air-gapped installs. Publishing may take several minutes.
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
          No app config found. Sync your app config before creating a bundle.
        </Text>
      )}
    </div>
  </Modal>
)
