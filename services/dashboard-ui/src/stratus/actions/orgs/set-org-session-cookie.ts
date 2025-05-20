'use server'

import { cookies } from 'next/headers'

export async function setOrgSessionCookie(orgId: string) {
  cookies().set('org-session', orgId)
}
