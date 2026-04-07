import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'

interface IShutdownRunnerModal extends Omit<IModal, 'onSubmit'> {
  showRunnerLabel?: boolean
  isPending: boolean
  error: TAPIError | null
  onSubmit: (force: boolean) => void
  onClose: () => void
}

export const ShutdownRunnerModal = ({
  showRunnerLabel,
  isPending,
  error,
  onSubmit,
  onClose,
  ...props
}: IShutdownRunnerModal) => {
  const label = showRunnerLabel ? 'Shutdown runner process' : 'Shutdown process'
  const [force, setForce] = useState(false)

  const handleClose = () => {
    setForce(false)
    onClose()
  }

  return (
    <Modal
      heading={
        <div className="flex flex-col gap-2">
          <Text
            flex
            className="gap-4"
            variant="h3"
            weight="strong"
            theme="warn"
          >
            <Icon variant="Power" size="24" />
            {label}?
          </Text>
        </div>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" />
            Shutting down
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="Power" />
            {label}
          </span>
        ),
        disabled: isPending,
        onClick: () => onSubmit(force),
        variant: 'primary' as const,
      }}
      onClose={handleClose}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to shutdown runner process.'}
          </Banner>
        ) : null}
        <div className="flex flex-col gap-4">
          <Text variant="base" weight="strong">
            Shutdown this runner process.
          </Text>
          <Text variant="body" theme="neutral" className="leading-relaxed max-w-md">
            This will terminate the container and restart the process. The runner
            will make a best effort to complete any queued jobs before shutting down.
          </Text>
          <ul className="flex flex-col gap-1 list-disc pl-6 text-sm text-cool-grey-700 dark:text-cool-grey-300">
            <li>Causes all jobs to queue while the process restarts</li>
            <li>Any new version updates will be applied on restart</li>
            <li>All local state will be refreshed</li>
          </ul>
          <div className="flex items-start">
            <CheckboxInput
              checked={force}
              onChange={(e) => setForce(e.target.checked)}
              className="mt-1.5"
              labelProps={{
                className:
                  'hover:!bg-transparent focus:!bg-transparent active:!bg-transparent !p-2 gap-4 max-w-none !items-start',
                labelText: (
                  <div className="flex flex-col gap-1">
                    <Text variant="base" weight="stronger">
                      Force shutdown
                    </Text>
                    <Text variant="subtext" theme="neutral">
                      Immediately shutdown the runner, terminating any in-flight
                      jobs. This has the potential for loss of state.
                    </Text>
                  </div>
                ),
              }}
            />
          </div>
        </div>
      </div>
    </Modal>
  )
}

interface IShutdownRunnerButton extends IButtonAsButton {
  showRunnerLabel?: boolean
  onOpenModal: () => void
}

export const ShutdownRunnerButton = ({
  showRunnerLabel,
  onOpenModal,
  ...props
}: IShutdownRunnerButton) => {
  const label = showRunnerLabel ? 'Shutdown runner process' : 'Shutdown process'

  return (
    <Button
      onClick={() => onOpenModal()}
      {...props}
    >
      {props?.isMenuButton ? null : <Icon variant="Power" />}
      {label}
      {props?.isMenuButton ? <Icon variant="Power" /> : null}
    </Button>
  )
}
