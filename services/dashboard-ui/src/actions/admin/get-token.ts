'use server'

import { getAccessToken } from '@/lib/auth-server'

export async function getToken() {
  const result = await getAccessToken()
  return { status: 200, result }
}