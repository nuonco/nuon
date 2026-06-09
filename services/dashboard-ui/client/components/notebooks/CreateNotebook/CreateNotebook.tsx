import { type FormEvent, useRef } from 'react'
import { Banner } from '@/components/common/Banner'
import { Input } from '@/components/common/form/Input'
import { Textarea } from '@/components/common/form/Textarea'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { ICreateNotebookBody } from '@/lib'

interface ICreateNotebookModal extends Omit<IModal, 'onSubmit'> {
  isPending: boolean
  error: any
  onSubmit: (body: ICreateNotebookBody) => void
}

export const CreateNotebookModal = ({
  isPending,
  error,
  onSubmit,
  ...props
}: ICreateNotebookModal) => {
  const formRef = useRef<HTMLFormElement>(null)

  const handleFormSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    const formData = new FormData(e.currentTarget)
    onSubmit({
      name: formData.get('name') as string,
      description: (formData.get('description') as string) || undefined,
    })
  }

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
          <Icon variant="NotebookIcon" size="24" />
          Create notebook
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? 'Creating...' : 'Create notebook',
        onClick: () => formRef.current?.requestSubmit(),
        disabled: isPending,
        variant: 'primary',
      }}
      {...props}
    >
      <form
        ref={formRef}
        onSubmit={handleFormSubmit}
        className="flex flex-col gap-4"
      >
        {error && <Banner theme="error">{error?.error}</Banner>}

        <label className="flex flex-col gap-1">
          <Text variant="label" weight="strong">
            Name
          </Text>
          <Input
            name="name"
            type="text"
            placeholder="e.g. Debug pods"
            required
            maxLength={255}
          />
        </label>

        <label className="flex flex-col gap-1">
          <Text variant="label" weight="strong">
            Description (optional)
          </Text>
          <Textarea
            name="description"
            placeholder="What this notebook is for"
            maxLength={2000}
            rows={3}
          />
        </label>
      </form>
    </Modal>
  )
}
