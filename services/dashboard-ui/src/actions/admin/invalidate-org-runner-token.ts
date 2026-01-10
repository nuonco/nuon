'use server'

import { getSession } from '@/lib/auth-server'
import { ADMIN_API_URL } from '@/configs/api'
import { getOrgRunner } from './get-org-runner'

export async function invalidateOrgRunnerToken(orgId: string) {
  const session = await getSession()
  const { user } = session || {}
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