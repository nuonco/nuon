'use server'

import type { TRunner } from '@/types'
import { ADMIN_API_URL } from '@/configs/api'

export async function getOrgRunner(orgId: string): Promise<TRunner> {
  const runner = await fetch(
    `${ADMIN_API_URL}/v1/orgs/${orgId}/admin-get-runner`
  ).then((r) => r.json())
  return runner
}