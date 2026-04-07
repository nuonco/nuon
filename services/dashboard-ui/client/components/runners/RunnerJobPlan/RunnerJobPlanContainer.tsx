import { useQuery } from '@tanstack/react-query'
import { type IButtonAsButton } from '@/components/common/Button'
import type { IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getRunnerJobPlan } from '@/lib'
import type { TRunnerJobPlan } from '@/types'
import {
  RunnerJobPlanModal as RunnerJobPlanModalComponent,
  RunnerJobPlanButton as RunnerJobPlanButtonComponent,
} from './RunnerJobPlan'

interface IRunnerJobPlan {
  runnerJobId: string
  buttonText?: string
  headingText?: string
}

export const RunnerJobPlanModal = ({
  runnerJobId,
  headingText = 'Runner job plan',
  ...props
}: Omit<IRunnerJobPlan, 'buttonText'> & IModal) => {
  const { org } = useOrg()

  const {
    data: plan,
    error,
    isLoading,
  } = useQuery<TRunnerJobPlan>({
    queryKey: ['runner-job-plan', org?.id, runnerJobId],
    queryFn: () => getRunnerJobPlan({ orgId: org.id, runnerJobId }),
    enabled: !!org?.id && !!runnerJobId,
  })

  return (
    <RunnerJobPlanModalComponent
      plan={plan}
      isLoading={isLoading}
      error={error}
      headingText={headingText}
      {...props}
    />
  )
}

export const RunnerJobPlanButton = ({
  runnerJobId,
  buttonText = 'View job plan',
  headingText,
  ...props
}: IRunnerJobPlan & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = (
    <RunnerJobPlanModal runnerJobId={runnerJobId} headingText={headingText} />
  )

  return (
    <RunnerJobPlanButtonComponent
      buttonText={buttonText}
      onOpenModal={() => addModal(modal)}
      {...props}
    />
  )
}
