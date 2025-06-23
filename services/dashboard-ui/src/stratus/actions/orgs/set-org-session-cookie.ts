'use server'

import { cookies } from 'next/headers'

export async function setOrgSessionCookie(orgId: string) {
  const cookieStore = await cookies()
  cookieStore.set('org-session', orgId)
}
