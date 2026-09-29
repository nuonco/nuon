import { api } from '@/lib/api'
import type { TAPIError, TOrg } from '@/types'

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

export async function createTrialOrg({
  attempts = TRIAL_ORG_ATTEMPTS,
}: { attempts?: number } = {}): Promise<TOrg> {
  let lastError: unknown
  for (let attempt = 0; attempt < attempts; attempt++) {
    const name = await fetchRandomName()
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
