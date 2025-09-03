import type { TWaitlist } from '@/types'
import { mutateData } from '@/utils'

export interface IJoinWaitlist {
  [org_name: string]: string
}

export async function joinWaitlist(data: IJoinWaitlist) {
  return mutateData<TWaitlist>({
    data,
    errorMessage:
      'Unable to add you to the Nuon waitlist, refresh the page and try again.',
    path: 'general/waitlist',
  })
}