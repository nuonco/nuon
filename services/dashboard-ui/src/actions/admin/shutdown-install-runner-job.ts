'use server'

import { auth0 } from '@/lib/auth'
import { ADMIN_API_URL } from '@/configs/api'
import { getInstallRunner } from './get-install-runner'

export async function shutdownInstallRunnerJob(installId: string) {
  const { user } = await auth0.getSession()
  const runner = await getInstallRunner(installId)

  try {
    const result = await fetch(
      `${ADMIN_API_URL}/v1/runners/${runner?.id}/shutdown-job`,
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
    throw new Error('Failed to kick off install runner shutdown job')
  }
}