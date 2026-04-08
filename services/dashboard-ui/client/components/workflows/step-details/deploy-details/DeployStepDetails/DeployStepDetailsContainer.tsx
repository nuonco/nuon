import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getDeploy } from '@/lib'
import type { TDeploy } from '@/types'
import type { IStepDetails } from '../../types'
import { DeployStepDetails } from './DeployStepDetails'

export const DeployStepDetailsContainer = ({ step }: IStepDetails) => {
  const { org } = useOrg()
  const {
    data: deploy,
    error,
    isLoading,
  } = useQuery<TDeploy>({
    queryKey: ['deploy', org?.id, step?.owner_id, step?.step_target_id],
    queryFn: () =>
      getDeploy({
        orgId: org.id,
        installId: step.owner_id,
        deployId: step.step_target_id,
      }),
    enabled: !!org?.id && !!step?.owner_id && !!step?.step_target_id,
  })

  return (
    <DeployStepDetails
      step={step}
      orgId={org?.id}
      deploy={deploy}
      error={error}
      isLoading={isLoading}
    />
  )
}
