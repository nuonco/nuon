import { api } from '@/lib/api'
import type { TAccount } from '@/types'

export const completeUserJourney = ({ journeyName }: { journeyName: string }) =>
  api<TAccount>({
    method: 'POST',
    path: `account/user-journeys/${journeyName}/complete`,
  })
