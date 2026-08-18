import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getInstallComponent } from '@/lib'
import type { IStepDetails } from '../types'
import { VerifyHealthStepDetails } from './VerifyHealthStepDetails'

interface IVerifyHealthStepDetailsContainer extends IStepDetails {}

export const VerifyHealthStepDetailsContainer = ({
  step,
}: IVerifyHealthStepDetailsContainer) => {
  const { org } = useOrg()
  const installId = step?.owner_id
  const installComponentId = step?.step_target_id

  // Component routes are keyed by component_id, not install_component_id.
  const { data: installComponent } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['install-component', org?.id, installId, installComponentId],
    queryFn: () =>
      getInstallComponent({
        orgId: org!.id,
        installId: installId!,
        componentId: installComponentId!,
      }),
    enabled: !!org?.id && !!installId && !!installComponentId,
  })

  return (
    <VerifyHealthStepDetails
      step={step}
      orgId={org?.id}
      componentId={installComponent?.component_id}
    />
  )
}
