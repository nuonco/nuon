'use server'

import { cookies } from 'next/headers'

export async function setOrgSessionCookie(orgId: string) {
  const c = await cookies()
  c.set('org-session', orgId)
}
