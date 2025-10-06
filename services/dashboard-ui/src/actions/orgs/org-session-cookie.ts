'use server'

import { cookies } from 'next/headers'

export async function setOrgCookie(orgId: string) {
  const cookieStore = await cookies()
  cookieStore.set('org_session', orgId, {
    path: '/',
    httpOnly: false,
    maxAge: 60 * 60 * 24 * 365,
    sameSite: 'lax',
  })
}

export async function getOrgIdFromCookie(): Promise<string> {
  const cookieStore = await cookies()
  return cookieStore.get('org_session')?.value
}
