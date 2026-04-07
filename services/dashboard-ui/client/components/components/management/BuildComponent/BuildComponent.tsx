import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TComponent, TAPIError } from '@/types'

export const BuildComponentButton = ({
  onClick,
  ...props
}: { onClick: () => void } & Omit<IButtonAsButton, 'onClick'>) => {
  return (
    <Button onClick={onClick} {...props}>
      {props?.isMenuButton ? null : <Icon variant="Hammer" />}
      Build component
      {props?.isMenuButton ? <Icon variant="Hammer" /> : null}
    </Button>
  )
}

interface IBuildComponentModal extends IModal {
  component: TComponent
  isLoading: boolean
  error?: TAPIError | null
  onBuild: () => void
}

export const BuildComponentModal = ({
  component,
  isLoading,
  error,
  onBuild,
  ...props
}: IBuildComponentModal) => {
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
          Build {component.name} component?
        </Text>
      }
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Building component
          </span>
        ) : (
          'Build component'
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
            {error?.error || 'Unable to build component'}
          </Banner>
        ) : null}
        <Text variant="base" weight="stronger">
          Are you sure you want to build {component.name}?
        </Text>
        <Text variant="base">
          This will start a build for the {component.name} component. The build
          process may take several minutes to complete.
        </Text>
        <Text variant="subtext" theme="neutral">
          You will be redirected to the build details page to monitor progress.
        </Text>
      </div>
    </Modal>
  )
}
