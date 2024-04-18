import type { TSandboxRunLogs } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export interface IGetSandboxRunLogs {
  installId: string
  orgId: string
  runId: string
}

export async function getSandboxRunLogs({
  orgId,
  installId,
  runId,
}: IGetSandboxRunLogs): Promise<TSandboxRunLogs> {
  const res = await fetch(
    `${API_URL}/v1/installs/${installId}/sandbox-run/${runId}/logs`,
    await getFetchOpts(orgId)
  )

  if (!res.ok) {
    throw new Error('Failed to fetch data')
  }

  return res.json()
}
