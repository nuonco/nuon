import { api } from '@/lib/api'
import type { TAccount } from '@/types'

// Clears each step's `complete` flag. Step metadata is left in place.
export const resetUserJourney = ({ journeyName }: { journeyName: string }) =>
  api<TAccount>({
    method: 'POST',
    path: `account/user-journeys/${journeyName}/reset`,
  })
