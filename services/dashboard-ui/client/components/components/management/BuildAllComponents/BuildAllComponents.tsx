import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'

export const BuildAllComponentsButton = ({
  onClick,
  ...props
}: { onClick: () => void } & Omit<IButtonAsButton, 'onClick'>) => {
  return (
    <Button onClick={onClick} {...props}>
      Build all components
      <Icon variant="Hammer" />
    </Button>
  )
}

interface IBuildAllComponentsModal extends IModal {
  appName: string
  isLoading: boolean
  error?: TAPIError | null
  onBuild: () => void
}

export const BuildAllComponentsModal = ({
  appName,
  isLoading,
  error,
  onBuild,
  ...props
}: IBuildAllComponentsModal) => {
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
          <Icon variant="Hammer" size="24" />
          Build all components for {appName}?
        </Text>
      }
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Building components
          </span>
        ) : (
          'Build components'
        ),
        disabled: isLoading,
        onClick: onBuild,
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-3 mb-6">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to build components'}
          </Banner>
        ) : null}
        <Text variant="base" weight="stronger">
          Are you sure you want to build all components?
        </Text>
        <Text variant="base">
          This will build all components in the application. This process may
          take several minutes to complete.
        </Text>
        <Text variant="subtext" theme="neutral">
          You can monitor the progress of each component build in the components
          table.
        </Text>
      </div>
    </Modal>
  )
}
