import { auth0 } from '@/lib/auth'

export async function getFetchOpts(
  orgId = '',
  headers = {},
  abortTimeout = 5000
): Promise<RequestInit> {
  const session = await auth0.getSession()
  return {
    cache: 'no-store',
    headers: {
      Authorization: `Bearer ${session?.tokenSet?.accessToken}`,
      'Content-Type': 'application/json',
      'X-Nuon-Org-ID': orgId,
      ...headers,
    },
    signal: AbortSignal.timeout(abortTimeout),
  }
}
