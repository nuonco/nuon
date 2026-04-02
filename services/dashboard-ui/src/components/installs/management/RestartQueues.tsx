'use client'

import { useState } from 'react'
import { restartInstallQueues } from '@/actions/admin/restart-install-queues'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useInstall } from '@/hooks/use-install'
import { useSurfaces } from '@/hooks/use-surfaces'

interface IRestartQueues {}

export const RestartQueuesModal = ({ ...props }: IRestartQueues & IModal) => {
  const { removeModal } = useSurfaces()
  const { install } = useInstall()
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleRestart = async () => {
    setIsLoading(true)
    setError(null)
    try {
      await restartInstallQueues(install.id)
      removeModal(props.modalId)
    } catch (err) {
      setError(
        err instanceof Error ? err.message : 'Failed to restart install queues'
      )
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <Modal
      heading={
        <Text
          className="inline-flex gap-4 items-center"
          variant="h3"
          weight="strong"
          theme="warning"
        >
          <Icon variant="ArrowsClockwise" size="24" />
          Restart install queues?
        </Text>
      }
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Restarting queues
          </span>
        ) : (
          'Restart queues'
        ),
        onClick: handleRestart,
        disabled: isLoading,
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-1">
        {error ? (
          <Banner theme="error">
            {error || 'An error happened, please refresh the page and try again.'}
          </Banner>
        ) : null}
        <Text variant="base" weight="strong">
          Are you sure you want to restart all queues for {install.name}?
        </Text>
        <Text variant="base">
          This will restart all Temporal queue workflows for the install. Use
          this when the install queue is stuck or unresponsive.
        </Text>
      </div>
    </Modal>
  )
}

export const RestartQueuesButton = ({
  ...props
}: IRestartQueues & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <RestartQueuesModal />

  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      Restart queues
      <Icon variant="ArrowsClockwise" />
    </Button>
  )
}
