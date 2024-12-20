import { API_URL, getFetchOpts } from '@/utils'

export async function postInstallReprovision({ installId, orgId }: {installId: string, orgId: string }): Promise<string> {
  const res = await fetch(`${API_URL}/v1/installs/${installId}/reprovision`, {
    ...(await getFetchOpts(orgId)),
    method: 'POST',
  })

  if (!res.ok) {
    throw new Error("Failed to kick off reprovision")
  }

  return res.json()
}
