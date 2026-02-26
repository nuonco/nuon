import { api } from '@/lib/api'

export type TRunActionBody = {
  action_workflow_config_id: string
  run_env_vars?: Record<string, string>
}

export async function runAction({
  body,
  installId,
  orgId,
}: {
  body: TRunActionBody
  installId: string
  orgId: string
}) {
  return api<string>({
    body,
    method: 'POST',
    orgId,
    path: `installs/${installId}/action-workflows/runs`,
  })
}
