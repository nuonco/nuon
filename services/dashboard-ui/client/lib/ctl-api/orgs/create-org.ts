import { api } from '@/lib/api'
import type { TAPIError, TOrg } from '@/types'
import { orgNameFromEmail } from './org-name-from-email'

export type TCreateOrgBody = {
  name: string
  use_sandbox_mode: boolean
  tags?: string[]
}

export const createOrg = ({ body }: { body: TCreateOrgBody }) =>
  api<TOrg>({
    body,
    method: 'POST',
    path: `orgs`,
  })

const TRIAL_ORG_ATTEMPTS = 5

export const fetchRandomName = async (): Promise<string> => {
  const res = await fetch('/api/random-name', { credentials: 'include' })
  if (!res.ok) throw new Error(`random-name returned ${res.status}`)
  const data = (await res.json()) as { name: string }
  return data.name
}

// Names the org after the user's work domain when there is one. A free email
// address, or a domain name another org already uses, gets a random name.
export async function createTrialOrg({
  attempts = TRIAL_ORG_ATTEMPTS,
  email,
}: { attempts?: number; email?: string } = {}): Promise<TOrg> {
  const domainName = orgNameFromEmail(email)
  let lastError: unknown
  for (let attempt = 0; attempt < attempts; attempt++) {
    const name = attempt === 0 && domainName ? domainName : await fetchRandomName()
    try {
      return await createOrg({
        body: { name, use_sandbox_mode: false, tags: ['Trial'] },
      })
    } catch (error) {
      if ((error as TAPIError)?.status !== 409) throw error
      lastError = error
    }
  }
  throw lastError
}
