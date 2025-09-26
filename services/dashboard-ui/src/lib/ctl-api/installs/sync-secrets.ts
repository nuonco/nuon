import { api } from '@/lib/api'

export type TSyncSecretsBody = {
  error_behavior?: "continue" | "abort" 
  plan_only: boolean
}

export async function syncSecrets({
  body,
  installId,
  orgId,
}: {
  body: TSyncSecretsBody
  installId: string
  orgId: string
}) {
  return api<string>({
    body,
    method: 'POST',
    orgId,
    path: `installs/${installId}/sync-secrets`,
  })
}
