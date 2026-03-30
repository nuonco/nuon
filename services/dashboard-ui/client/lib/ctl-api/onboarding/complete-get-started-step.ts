import { api } from '@/lib/api'
import type { TOnboarding } from '@/types'

export const completeGetStartedStep = ({ orgId }: { orgId: string }) =>
  api<TOnboarding>({
    method: 'POST',
    orgId,
    path: `onboarding/current/steps/get-started`,
  })
