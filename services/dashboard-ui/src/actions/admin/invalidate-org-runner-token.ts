'use server'

import { auth0 } from '@/lib/auth'
import { ADMIN_API_URL } from '@/configs/api'
import { getOrgRunner } from './get-org-runner'

export async function invalidateOrgRunnerToken(orgId: string) {
  const { user } = await auth0.getSession()
  const runner = await getOrgRunner(orgId)

  try {
    const result = await fetch(
      `${ADMIN_API_URL}/v1/runners/${runner.id}/invalidate-service-account-token`,
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
    throw new Error('Failed to invalidate service account token')
  }
}