import { useState } from 'react'
import type { ReactNode } from 'react'
import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'

interface IDeployAllComponentsModal extends Omit<IModal, 'onSubmit'> {
  installName: string
  isPending: boolean
  isKickedOff: boolean
  error?: { error?: string } | null
  onSubmit: (params: { role: string }) => void
  roleSelector: (props: {
    value: string
    onChange: (value: string) => void
  }) => ReactNode
}

export const DeployAllComponentsModal = ({
  installName,
  isPending,
  isKickedOff,
  error,
  onSubmit,
  roleSelector,
  ...props
}: IDeployAllComponentsModal) => {
  const [selectedRole, setSelectedRole] = useState<string>('')

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
          <Icon variant="CloudArrowUpIcon" size="24" />
          Deploy all components to {installName}?
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Starting deployments
          </span>
        ) : (
          'Deploy all components'
        ),
        disabled: isKickedOff || isPending,
        onClick: () => onSubmit({ role: selectedRole }),
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-3 mb-6">
        {(error as any)?.error ? (
          <Banner theme="error">
            {(error as any)?.error || 'Unable to deploy components'}
          </Banner>
        ) : null}
        <Text variant="base">
          This will deploy the latest build of each component to your install.
        </Text>
        {roleSelector({
          value: selectedRole,
          onChange: setSelectedRole,
        })}
      </div>
    </Modal>
  )
}
