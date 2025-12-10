'use server'

import { auth0 } from '@/lib/auth'

export async function getToken() {
  const result = await auth0.getAccessToken()
  return { status: 200, result }
}