'use server'

import { getSession } from '@/lib/auth-server'
import { ADMIN_API_URL } from '@/configs/api'

export async function reprovisionApp(appId: string) {
  const session = await getSession()
  const { user } = session || {}

  try {
    const result = await fetch(
      `${ADMIN_API_URL}/v1/apps/${appId}/admin-reprovision`,
      {
        method: 'POST',
        body: '{}',
        headers: {
          'Content-Type': 'application/json',
          'X-Nuon-Admin-Email': user?.email,
        },
      }
    ).then((r) => r.json())
    return { status: 201, result }
  } catch (error) {
    throw new Error('Failed to kick off app reprovision')
  }
}