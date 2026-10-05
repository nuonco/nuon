import { api } from '@/lib/api'
import type { TAccount } from '@/types'

export type TCreateUserJourneyBody = {
  name: string
  title: string
  steps: { name: string; title: string }[]
}

export const createUserJourney = ({ body }: { body: TCreateUserJourneyBody }) =>
  api<TAccount>({
    method: 'POST',
    path: 'account/user-journeys',
    body,
  })
