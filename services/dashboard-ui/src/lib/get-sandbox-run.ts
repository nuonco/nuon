import type { TSandboxRun } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export interface IGetSandboxRun {
  installId: string
  orgId: string
  runId: string
}

export async function getSandboxRun({
  orgId,
  installId,
  runId,
}: IGetSandboxRun): Promise<TSandboxRun> {
  const res = await fetch(
    `${API_URL}/v1/installs/${installId}/sandbox-runs`,
    await getFetchOpts(orgId)
  )

  if (!res.ok) {
    throw new Error('Failed to fetch sandbox run')
  }

  return res
    .json()
    .then((runs: Array<TSandboxRun>) => runs.find((r) => r?.id === runId))
}
