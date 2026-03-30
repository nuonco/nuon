import { api } from '@/lib/api'
import type { TOnboarding } from '@/types'

export const completeDeployStep = ({ orgId }: { orgId: string }) =>
  api<TOnboarding>({
    method: 'POST',
    orgId,
    path: `onboarding/current/steps/deploy`,
  })
