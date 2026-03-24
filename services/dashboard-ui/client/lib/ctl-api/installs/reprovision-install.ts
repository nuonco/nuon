import { api } from '@/lib/api'

export type TReprovisionInstallBody = {
  plan_only: boolean
  role?: string
}

export async function reprovisionInstall({
  body,
  installId,
  orgId,
}: {
  body: TReprovisionInstallBody
  installId: string
  orgId: string
}) {
  return api<string>({
    withHeaders: true,
    body,
    method: 'POST',
    orgId,
    path: `installs/${installId}/reprovision`,
  })
}
