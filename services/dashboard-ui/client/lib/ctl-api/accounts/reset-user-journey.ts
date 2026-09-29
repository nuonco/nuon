import { api } from '@/lib/api'
import type { TAccount } from '@/types'

export const resetUserJourney = ({ journeyName }: { journeyName: string }) =>
  api<TAccount>({
    method: 'POST',
    path: `account/user-journeys/${journeyName}/reset`,
  })
