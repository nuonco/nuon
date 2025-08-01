import type { TWaitlist } from '@/types'
import { API_URL, getFetchOpts, mutateData } from '@/utils'

export async function getAPIVersion(): Promise<{
  git_ref: string
  version: string
}> {
  const data = await fetch(`${API_URL}/version`, await getFetchOpts()).catch(
    (error) => {
      console.error(error)
      return {
        ok: false,
      } as Response
    }
  )

  if (!data?.ok) {
    throw new Error('Failed to fetch api version')
  }

  return data && data?.json()
}

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
