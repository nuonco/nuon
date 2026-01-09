import { getSession } from '@/lib/auth-server'

export async function getFetchOpts(
  orgId = '',
  headers = {},
  abortTimeout = 5000
): Promise<RequestInit> {
  const session = await getSession()
  return {
    cache: 'no-store',
    headers: {
      Authorization: `Bearer ${session?.accessToken}`,
      'Content-Type': 'application/json',
      'X-Nuon-Org-ID': orgId,
      ...headers,
    },
    signal: AbortSignal.timeout(abortTimeout),
  }
}
