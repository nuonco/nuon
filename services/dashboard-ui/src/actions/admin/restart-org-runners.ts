'use server'

import { auth0 } from '@/lib/auth'
import { ADMIN_API_URL } from '@/configs/api'

export async function restartOrgRunners(orgId: string) {
  const { user } = await auth0.getSession()

  try {
    const result = await fetch(
      `${ADMIN_API_URL}/v1/orgs/${orgId}/admin-restart-runners`,
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
    throw new Error('Failed to restart the org children event loops')
  }
}