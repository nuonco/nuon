import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'

interface IBuildSandboxModal extends Omit<IModal, 'onSubmit'> {
  appName: string
  isPending: boolean
  error: any
  onSubmit: () => void
}

export const BuildSandboxModal = ({
  appName,
  isPending,
  error,
  onSubmit,
  ...props
}: IBuildSandboxModal) => {
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
          Build sandbox?
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Building sandbox
          </span>
        ) : (
          'Build sandbox'
        ),
        disabled: isPending,
        onClick: onSubmit,
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-3 mb-6">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to start sandbox build'}
          </Banner>
        ) : null}
        <Text variant="base" weight="stronger">
          Are you sure you want to build the sandbox for {appName}?
        </Text>
        <Text variant="base">
          This will start a sandbox build. The build process may take several
          minutes to complete.
        </Text>
      </div>
    </Modal>
  )
}
