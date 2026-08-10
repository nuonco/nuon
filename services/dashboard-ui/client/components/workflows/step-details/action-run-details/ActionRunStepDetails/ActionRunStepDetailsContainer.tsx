import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { isTerminalStatusV2 } from '@/hooks/use-sse-resource-query'
import { getInstallActionRun } from '@/lib'
import type { TInstallActionRun } from '@/types'
import type { IActionRunDetails } from '../types'
import { ActionRunStepDetails } from './ActionRunStepDetails'

export const ActionRunStepDetailsContainer = ({ step }: IActionRunDetails) => {
  const { org } = useOrg()

  const {
    data: actionRun,
    error,
    isLoading,
  } = useQuery<TInstallActionRun>({
    queryKey: ['action-run', org?.id, step?.owner_id, step?.step_target_id],
    queryFn: () =>
      getInstallActionRun({
        orgId: org!.id,
        installId: step!.owner_id,
        runId: step!.step_target_id,
      }),
    enabled: !!org?.id && !!step?.owner_id && !!step?.step_target_id,
    refetchInterval: (query) =>
      isTerminalStatusV2(query.state.data) ? false : 1_500,
  })

  const createdBy = actionRun?.created_by

  return (
    <ActionRunStepDetails
      step={step}
      actionRun={actionRun}
      createdBy={createdBy}
      error={error}
      isLoading={isLoading}
    />
  )
}
