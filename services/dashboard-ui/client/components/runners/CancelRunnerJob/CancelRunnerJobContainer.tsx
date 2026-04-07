import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { cancelRunnerJob } from '@/lib'
import type { TAPIError, TRunnerJob } from '@/types'
import {
  CancelRunnerJobModal as CancelRunnerJobModalComponent,
  CancelRunnerJobButton as CancelRunnerJobButtonComponent,
  cancelJobOptions,
  type TCancelJobType,
} from './CancelRunnerJob'

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
}: ICancelRunnerJob & Omit<IModal, 'onSubmit'>) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const cancelJobData = cancelJobOptions[jobType]

  const { mutate, isPending: isLoading, error } = useMutation<unknown, TAPIError>({
    mutationFn: () => cancelRunnerJob({ orgId: org.id, runnerJobId: runnerJob.id }),
    onSuccess: () => {
      addToast(
        <Toast heading={`${cancelJobData.buttonText} successful.`} theme="success">
          <Text>Successfully cancelled the {jobType} job.</Text>
        </Toast>
      )
      removeModal(props.modalId)
      onSuccess?.()
    },
    onError: (err: { error?: string }) => {
      addToast(
        <Toast heading={`${cancelJobData.buttonText} failed.`} theme="error">
          <Text>
            There was an error while trying to cancel {jobType} job{' '}
            {runnerJob.id}.
          </Text>
          <Text>{err?.error || 'Unknown error occurred.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <CancelRunnerJobModalComponent
      runnerJob={runnerJob}
      jobType={jobType}
      isPending={isLoading}
      error={error}
      onSubmit={() => mutate()}
      {...props}
    />
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

  const modal = (
    <CancelRunnerJobModal
      runnerJob={runnerJob}
      jobType={jobType}
      onSuccess={() => setHasBeenCanceled(true)}
    />
  )

  return (
    <CancelRunnerJobButtonComponent
      runnerJob={runnerJob}
      jobType={jobType}
      isDisabled={isDisabled}
      hasBeenCanceled={hasBeenCanceled}
      onOpenModal={() => addModal(modal)}
      {...props}
    />
  )
}
