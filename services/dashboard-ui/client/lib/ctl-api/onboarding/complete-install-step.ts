import { api } from '@/lib/api'
import type { TCompleteInstallStepRequest, TOnboarding } from '@/types'

export const completeInstallStep = ({
  body,
  orgId,
}: {
  body: TCompleteInstallStepRequest
  orgId: string
}) =>
  api<TOnboarding>({
    body,
    method: 'POST',
    orgId,
    path: `onboarding/current/steps/install`,
  })
