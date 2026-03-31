import { api } from '@/lib/api'
import type { TCompleteYourStackStepRequest, TOnboarding } from '@/types'

export const completeYourStackStep = ({
  body,
  orgId,
}: {
  body: TCompleteYourStackStepRequest
  orgId: string
}) =>
  api<TOnboarding>({
    body,
    method: 'POST',
    orgId,
    path: `onboarding/current/steps/your-stack`,
  })
