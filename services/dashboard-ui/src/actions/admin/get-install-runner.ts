'use server'

import type { TRunner } from '@/types'
import { ADMIN_API_URL } from '@/configs/api'

export async function getInstallRunner(installId: string): Promise<TRunner> {
  const runner = await fetch(
    `${ADMIN_API_URL}/v1/installs/${installId}/admin-get-runner`
  ).then((r) => r.json())
  return runner
}