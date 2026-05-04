import { useMutation, useQueryClient } from '@tanstack/react-query'
import { type IButtonAsButton } from '@/components/common/Button'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { retryNowWorkflowStep } from '@/lib'
import type { TAPIError, TWorkflowStep } from '@/types'
import { RetryNowButton } from './RetryNow'

interface IRetryNow {
  step: TWorkflowStep
}

export const RetryNowButtonContainer = ({
  step,
  ...props
}: IRetryNow & IButtonAsButton) => {
  const { org } = useOrg()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const { mutate: execute, isPending } = useMutation<unknown, TAPIError>({
    mutationFn: () =>
      retryNowWorkflowStep({
        orgId: org.id,
        workflowId: step.install_workflow_id,
        stepId: step.id,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['workflow-steps'] })
      queryClient.invalidateQueries({ queryKey: ['active-workflows'] })
    },
    onError: (err) => {
      addToast(
        <Toast heading="Failed to retry now" theme="error">
          <Text>{err?.error || 'Unknown error occurred.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <RetryNowButton
      retryNotBeforeAt={step.retry_not_before_at}
      isPending={isPending}
      onTrigger={() => execute()}
      {...props}
    />
  )
}
