import { type FormEvent, useRef } from 'react'
import { Input } from '@/components/common/form/Input'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'

interface ICreateAppModal extends Omit<IModal, 'onSubmit'> {
  isSubmitting: boolean
  onSubmit: (body: { name: string }) => void
}

export const CreateAppModal = ({
  isSubmitting,
  onSubmit,
  ...props
}: ICreateAppModal) => {
  const formRef = useRef<HTMLFormElement>(null)

  const handleFormSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    const name = (new FormData(e.currentTarget).get('name') as string).trim()
    if (!name) return
    onSubmit({ name })
  }

  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon variant="AppWindowIcon" size="24" />
          Create app
        </Text>
      }
      primaryActionTrigger={{
        children: isSubmitting ? 'Creating app' : 'Create app',
        onClick: () => formRef.current?.requestSubmit(),
        disabled: isSubmitting,
        variant: 'primary',
      }}
      {...props}
    >
      <form
        ref={formRef}
        onSubmit={handleFormSubmit}
        className="flex flex-col gap-4"
      >
        <Input
          name="name"
          type="text"
          labelProps={{ labelText: 'App name' }}
          placeholder="my-app"
          required
          maxLength={255}
          disabled={isSubmitting}
        />
      </form>
    </Modal>
  )
}
