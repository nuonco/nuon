import { type FormEvent, useRef } from 'react'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Modal, type IModal } from '@/components/surfaces/Modal'

interface IRegisterAirgapInstallModal extends Omit<IModal, 'onSubmit'> {
  bundleId?: string
  isPending: boolean
  error?: { error?: string } | null
  onSubmit: (body: { name: string }) => void
}

export const RegisterAirgapInstallModal = ({
  bundleId,
  isPending,
  error,
  onSubmit,
  ...props
}: IRegisterAirgapInstallModal) => {
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
        <Text flex className="gap-4" variant="h3" weight="strong" theme="info">
          <Icon variant="CloudSlashIcon" size="24" />
          Register install
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Registering install
          </span>
        ) : (
          'Register install'
        ),
        disabled: isPending,
        onClick: () => formRef.current?.requestSubmit(),
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-3 mb-6">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to register the install.'}
          </Banner>
        ) : null}
        <Text variant="base">
          Track an air-gapped delivery of this bundle. This only creates a
          record of who the bundle was delivered to — nothing is provisioned,
          and the customer&apos;s environment never connects back to Nuon.
        </Text>
        {bundleId ? (
          <Text variant="subtext" theme="neutral" flex className="gap-2">
            Bundle
            <Badge variant="code" size="sm">
              {bundleId}
            </Badge>
          </Text>
        ) : null}
        <form
          ref={formRef}
          onSubmit={handleFormSubmit}
          className="flex flex-col gap-4"
        >
          <Input
            name="name"
            type="text"
            labelProps={{ labelText: 'Install name' }}
            placeholder="acme-prod"
            required
            maxLength={255}
            disabled={isPending}
          />
        </form>
      </div>
    </Modal>
  )
}
