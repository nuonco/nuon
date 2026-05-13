import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'
import type { TRunnerJob } from '@/types'

export type TCancelJobType =
  | 'build'
  | 'deploy'
  | 'sandbox-run'
  | 'workflow-run'
  | 'sandbox'
  | 'actions'
  | 'sync'
  | 'operations'
  | 'runner'
  | 'health-checks'

type TCancelJobData = {
  buttonText: string
  confirmHeading: string
  confirmMessage: string
}

export const cancelJobOptions: Record<TCancelJobType, TCancelJobData> = {
  build: {
    buttonText: 'Cancel build',
    confirmHeading: 'Cancel component build?',
    confirmMessage: 'Are you sure you want to cancel this component build?',
  },
  deploy: {
    buttonText: 'Cancel deploy',
    confirmHeading: 'Cancel component deployment?',
    confirmMessage: 'Are you sure you want to cancel this component deployment?',
  },
  sync: {
    buttonText: 'Cancel sync',
    confirmHeading: 'Cancel component sync?',
    confirmMessage: 'Are you sure you want to cancel this component sync?',
  },
  'sandbox-run': {
    buttonText: 'Cancel sandbox job',
    confirmHeading: 'Cancel sandbox job?',
    confirmMessage: 'Are you sure you want to cancel this sandbox job?',
  },
  sandbox: {
    buttonText: 'Cancel sandbox job',
    confirmHeading: 'Cancel sandbox job?',
    confirmMessage: 'Are you sure you want to cancel this sandbox job?',
  },
  'workflow-run': {
    buttonText: 'Cancel action',
    confirmHeading: 'Cancel action workflow?',
    confirmMessage: 'Are you sure you want to cancel this action workflow?',
  },
  actions: {
    buttonText: 'Cancel action',
    confirmHeading: 'Cancel action workflow?',
    confirmMessage: 'Are you sure you want to cancel this action workflow?',
  },
  operations: {
    buttonText: 'Cancel action',
    confirmHeading: 'Cancel shut down?',
    confirmMessage: 'Are you sure you want to cancel this shut down job?',
  },
  runner: {
    buttonText: 'Cancel action',
    confirmHeading: 'Cancel runner job?',
    confirmMessage: 'Are you sure you want to cancel this runner job?',
  },
  'health-checks': {
    buttonText: 'Cancel action',
    confirmHeading: 'Cancel health check?',
    confirmMessage: 'Are you sure you want to cancel this health check job?',
  },
}

interface ICancelRunnerJobModal extends Omit<IModal, 'onSubmit'> {
  runnerJob: TRunnerJob
  jobType: TCancelJobType
  isPending: boolean
  error: TAPIError | null
  onSubmit: () => void
}

export const CancelRunnerJobModal = ({
  runnerJob,
  jobType,
  isPending,
  error,
  onSubmit,
  ...props
}: ICancelRunnerJobModal) => {
  const cancelJobData = cancelJobOptions[jobType]

  return (
    <Modal
      heading={
        <Text
          flex
          className="gap-4"
          variant="h3"
          weight="strong"
          theme="error"
        >
          <Icon variant="WarningIcon" size="24" />
          {cancelJobData.confirmHeading}
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" className="animate-spin" />
            Canceling job
          </span>
        ) : (
          cancelJobData.buttonText
        ),
        disabled: isPending,
        onClick: () => onSubmit(),
        variant: 'danger',
      }}
      {...props}
    >
      <div className="flex flex-col gap-1">
        {error ? (
          <Banner theme="error">
            {error?.error ||
              'An error happned, please refresh the page and try again.'}
          </Banner>
        ) : null}
        <Text variant="base" weight="strong">
          {cancelJobData.confirmMessage}
        </Text>
        <Text variant="base">
          Once a runner job is canceled, it cannot be restarted. The job will stop
          immediately and any in-progress work will be lost.
        </Text>
      </div>
    </Modal>
  )
}

interface ICancelRunnerJobButton extends IButtonAsButton {
  runnerJob: TRunnerJob
  jobType: TCancelJobType
  isDisabled?: boolean
  hasBeenCanceled?: boolean
  onOpenModal: () => void
}

export const CancelRunnerJobButton = ({
  runnerJob,
  jobType,
  isDisabled,
  hasBeenCanceled,
  onOpenModal,
  ...props
}: ICancelRunnerJobButton) => {
  const cancelJobData = cancelJobOptions[jobType]

  return (
    <Button
      variant="danger"
      disabled={isDisabled || hasBeenCanceled}
      onClick={() => onOpenModal()}
      {...props}
    >
      {hasBeenCanceled ? 'Canceled' : cancelJobData.buttonText}
    </Button>
  )
}
