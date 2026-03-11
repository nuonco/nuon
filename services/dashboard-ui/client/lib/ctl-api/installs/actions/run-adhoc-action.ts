import { api } from '@/lib/api'

export type TRunAdhocActionBody = {
  command?: string
  env_vars?: Record<string, string>
  inline_contents?: string
  name?: string
  role?: string
  timeout?: number
}

export async function runAdhocAction({
  body,
  installId,
  orgId,
}: {
  body: TRunAdhocActionBody
  installId: string
  orgId: string
}) {
  return api<string>({
    withHeaders: true,
    body,
    method: 'POST',
    orgId,
    path: `installs/${installId}/actions/adhoc-run`,
  })
}
