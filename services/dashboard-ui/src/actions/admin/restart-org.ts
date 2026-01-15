'use server'

import { getSession } from '@/lib/auth-server'
import { ADMIN_API_URL } from '@/configs/api'

export async function restartOrg(orgId: string) {
  const session = await getSession()
  const { user } = session || {}

  try {
    const result = await fetch(
      `${ADMIN_API_URL}/v1/orgs/${orgId}/admin-restart`,
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
    throw new Error('Failed to restart the org event loop')
  }
}