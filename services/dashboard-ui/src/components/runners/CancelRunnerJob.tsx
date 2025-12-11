'use client'

import { useState } from 'react'
import { usePathname } from 'next/navigation'
import { cancelRunnerJob } from '@/actions/runners/cancel-runner-job'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useServerAction } from '@/hooks/use-server-action'
import { useServerActionToast } from '@/hooks/use-server-action-toast'
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

const cancelJobOptions: Record<TCancelJobType, TCancelJobData> = {
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

interface ICancelRunnerJob {
  runnerJob: TRunnerJob
  jobType: TCancelJobType
  isDisabled?: boolean
  onSuccess?: () => void
}

export const CancelRunnerJobModal = ({
  runnerJob,
  jobType,
  onSuccess,
  ...props
}: ICancelRunnerJob & IModal) => {
  const path = usePathname()
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const cancelJobData = cancelJobOptions[jobType]
  
  const { data, error, isLoading, execute } = useServerAction({
    action: cancelRunnerJob,
  })

  useServerActionToast({
    data,
    error,
    errorContent: (
      <>
        <Text>
          There was an error while trying to cancel {jobType} job{' '}
          {runnerJob.id}.
        </Text>
        <Text>{error?.error || 'Unknown error occurred.'}</Text>
      </>
    ),
    errorHeading: `${cancelJobData.buttonText} failed.`,
    onSuccess: () => {
      removeModal(props.modalId)
      onSuccess?.()
    },
    successContent: (
      <Text>Successfully cancelled the {jobType} job.</Text>
    ),
    successHeading: `${cancelJobData.buttonText} successful.`,
  })

  return (
    <Modal
      heading={
        <Text
          className="inline-flex gap-4 items-center"
          variant="h3"
          weight="strong"
          theme="error"
        >
          <Icon variant="Warning" size="24" />
          {cancelJobData.confirmHeading}
        </Text>
      }
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" className="animate-spin" />
            Canceling job
          </span>
        ) : (
          cancelJobData.buttonText
        ),
        disabled: isLoading,
        onClick: () => {
          execute({ 
            orgId: org.id, 
            path, 
            runnerJobId: runnerJob.id 
          })
        },
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

export const CancelRunnerJobButton = ({
  runnerJob,
  jobType,
  isDisabled,
  ...props
}: ICancelRunnerJob & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const [hasBeenCanceled, setHasBeenCanceled] = useState(false)
  const cancelJobData = cancelJobOptions[jobType]
  
  const modal = (
    <CancelRunnerJobModal 
      runnerJob={runnerJob} 
      jobType={jobType}
      onSuccess={() => setHasBeenCanceled(true)}
    />
  )

  return (
    <Button
      variant="danger"
      disabled={isDisabled || hasBeenCanceled}
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      {hasBeenCanceled ? 'Canceled' : cancelJobData.buttonText}
    </Button>
  )
}