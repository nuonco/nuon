import { api } from '@/lib/api'

export type TReprovisionSandboxBody = {
  plan_only: boolean
  skip_components?: boolean
}

export async function reprovisionSandbox({
  body,
  installId,
  orgId,
}: {
  body: TReprovisionSandboxBody
  installId: string
  orgId: string
}) {
  return api<string>({
    body,
    method: 'POST',
    orgId,
    path: `installs/${installId}/reprovision-sandbox`,
  })
}
