'use server'

import { getSession } from '@/lib/auth-server'
import { ADMIN_API_URL } from '@/configs/api'
import { getInstallRunner } from './get-install-runner'

export async function shutdownInstallRunnerJob(installId: string) {
  const session = await getSession()
  const { user } = session || {}
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