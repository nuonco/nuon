import { api } from '@/lib/api'

export type TBootstrapTokenResponse = {
  token: string
  expires_at: string
}

export async function createRunnerBootstrapToken({
  installId,
  orgId,
}: {
  installId: string
  orgId: string
}) {
  return api<TBootstrapTokenResponse>({
    method: 'POST',
    orgId,
    path: `installs/${installId}/runner-bootstrap-token`,
  })
}
