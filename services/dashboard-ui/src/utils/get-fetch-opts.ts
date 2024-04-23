import { getSession } from '@auth0/nextjs-auth0'

export async function getFetchOpts(orgId = ""): Promise<RequestInit> {
  const session = await getSession()
  return {
    cache: "no-store",
    headers: {
      Authorization: `Bearer ${session?.accessToken}`,
      'Content-Type': 'application/json',
      'X-Nuon-Org-ID': orgId,
    },    
  }
}
