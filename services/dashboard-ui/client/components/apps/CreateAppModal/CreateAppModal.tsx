import { useState } from 'react'
import { Input } from '@/components/common/form/Input'
import { Modal, type IModal } from '@/components/surfaces/Modal'

interface ICreateAppModal extends Omit<IModal, 'onSubmit'> {
  isSubmitting: boolean
  onSubmit: (body: { name: string }) => void
  onCancel: () => void
}

export const CreateAppModal = ({
  isSubmitting,
  onSubmit,
  onCancel,
  ...props
}: ICreateAppModal) => {
  const [name, setName] = useState('')

  return (
    <Modal
      heading="Create app"
      size="sm"
      primaryActionTrigger={{
        children: 'Create',
        disabled: !name.trim() || isSubmitting,
        onClick: () => onSubmit({ name: name.trim() }),
      }}
      secondaryActionTrigger={{
        children: 'Cancel',
        onClick: onCancel,
      }}
      showFooter
      {...props}
    >
      <div className="flex flex-col gap-4 p-6">
        <Input
          labelProps={{ labelText: 'App name' }}
          placeholder="my-app"
          value={name}
          onChange={(e) => setName(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === 'Enter' && name.trim()) {
              onSubmit({ name: name.trim() })
            }
          }}
          autoFocus
        />
      </div>
    </Modal>
  )
}
