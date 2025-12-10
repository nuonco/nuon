'use server'

import { auth0 } from '@/lib/auth'
import { ADMIN_API_URL } from '@/configs/api'

export async function teardownInstallComponents(installId: string) {
  const { user } = await auth0.getSession()

  try {
    const result = await fetch(
      `${ADMIN_API_URL}/v1/installs/${installId}/admin-teardown-components`,
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
    throw new Error('Failed to teardown install components')
  }
}