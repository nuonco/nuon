import { api } from '@/lib/api'
import type { TCompleteOrganizationStepRequest, TOnboarding } from '@/types'

export const completeOrganizationStep = ({
  body,
}: {
  body: TCompleteOrganizationStepRequest
}) =>
  api<TOnboarding>({
    body,
    method: 'POST',
    path: `onboarding/current/steps/organization`,
  })
