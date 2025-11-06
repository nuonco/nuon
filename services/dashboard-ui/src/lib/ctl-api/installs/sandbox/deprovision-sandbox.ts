import { api } from '@/lib/api'

export type TDeprovisionSandboxBody = {
  error_behavior?: 'continue' | 'abort'
  plan_only: boolean
}

export async function deprovisionSandbox({
  body,
  installId,
  orgId,
}: {
  body: TDeprovisionSandboxBody
  installId: string
  orgId: string
}) {
  return api<string>({
    body,
    method: 'POST',
    orgId,
    path: `installs/${installId}/deprovision-sandbox`,
  })
}
