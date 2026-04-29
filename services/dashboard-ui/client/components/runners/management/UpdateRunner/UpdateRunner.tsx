import { useRef, useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Input } from '@/components/common/form/Input'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'

interface IUpdateRunnerModal extends Omit<IModal, 'onSubmit'> {
  isPending: boolean
  error: TAPIError | null
  onSubmit: (tag: string) => void
  onClose: () => void
  modalHeading?: string
  inputLabel?: string
  inputPlaceholder?: string
  submitLabel?: string
}

export const UpdateRunnerModal = ({
  isPending,
  error,
  onSubmit,
  onClose,
  modalHeading = 'Update runner version',
  inputLabel = 'Enter the runner tag you\'d like to update to.',
  inputPlaceholder = 'runner tag',
  submitLabel = 'Update runner version',
  ...props
}: IUpdateRunnerModal) => {
  const formRef = useRef<HTMLFormElement>(null)
  const [tag, setTag] = useState('')

  const canUpdate = tag.trim().length > 0 && !isPending

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    onSubmit(tag)
  }

  const handleFormSubmit = () => {
    if (formRef.current) {
      formRef.current.requestSubmit()
    }
  }

  const handleClose = () => {
    setTag('')
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
            theme="info"
          >
            <Icon variant="ArrowsCounterClockwise" size="24" />
            {modalHeading}
          </Text>
        </div>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" />
            Updating runner
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="ArrowsCounterClockwise" />
            {submitLabel}
          </span>
        ),
        disabled: !canUpdate,
        onClick: handleFormSubmit,
        variant: 'primary' as const,
      }}
      onClose={handleClose}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to update runner.'}
          </Banner>
        ) : null}
        <form ref={formRef} onSubmit={handleSubmit}>
          <div className="flex flex-col gap-4">
            <div className="flex flex-col gap-2">
              <Text variant="base" weight="stronger">
                {inputLabel}
              </Text>
              <Input
                id="runner-tag"
                placeholder={inputPlaceholder}
                type="text"
                value={tag}
                onChange={(e) => setTag(e.target.value)}
                required
              />
            </div>
          </div>
        </form>
      </div>
    </Modal>
  )
}

interface IUpdateRunnerButton extends IButtonAsButton {
  onOpenModal: () => void
  label?: string
}

export const UpdateRunnerButton = ({
  onOpenModal,
  label = 'Update runner version',
  ...props
}: IUpdateRunnerButton) => {
  return (
    <Button
      onClick={() => onOpenModal()}
      {...props}
    >
      {props?.isMenuButton ? null : <Icon variant="ArrowsCounterClockwise" />}
      {label}
      {props?.isMenuButton ? <Icon variant="ArrowsCounterClockwise" /> : null}
    </Button>
  )
}
