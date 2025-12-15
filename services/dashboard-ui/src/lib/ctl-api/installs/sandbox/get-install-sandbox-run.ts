import { api } from '@/lib/api'
import type { TSandboxRun } from '@/types'

export const getInstallSandboxRun = ({
  runId,
  orgId,
}: {
  runId: string
  orgId: string
}) =>
  api<TSandboxRun>({
    path: `installs/sandbox-runs/${runId}`,
    orgId,
  })
