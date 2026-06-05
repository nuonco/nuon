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
}

export const cancelJobOptions: Record<TCancelJobType, TCancelJobData> = {
  build: {
    buttonText: 'Cancel build',
    confirmHeading: 'Cancel component build?',
  },
  deploy: {
    buttonText: 'Cancel deploy',
    confirmHeading: 'Cancel component deployment?',
  },
  sync: {
    buttonText: 'Cancel sync',
    confirmHeading: 'Cancel component sync?',
  },
  'sandbox-run': {
    buttonText: 'Cancel sandbox job',
    confirmHeading: 'Cancel sandbox job?',
  },
  sandbox: {
    buttonText: 'Cancel sandbox job',
    confirmHeading: 'Cancel sandbox job?',
  },
  'workflow-run': {
    buttonText: 'Cancel action',
    confirmHeading: 'Cancel action workflow?',
  },
  actions: {
    buttonText: 'Cancel action',
    confirmHeading: 'Cancel action workflow?',
  },
  operations: {
    buttonText: 'Cancel action',
    confirmHeading: 'Cancel shut down?',
  },
  runner: {
    buttonText: 'Cancel action',
    confirmHeading: 'Cancel runner job?',
  },
  'health-checks': {
    buttonText: 'Cancel action',
    confirmHeading: 'Cancel health check?',
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
              'Something went wrong. Try refreshing the page.'}
          </Banner>
        ) : null}
        <Text variant="base">
          Once canceled, the job cannot be restarted. It will stop immediately
          and any in-progress work will be lost.
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
