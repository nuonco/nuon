import { TReadme } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export interface IGetInstallReadme {
  installId: string
  orgId: string
}

export async function getInstallReadme({
  installId,
  orgId,
}: IGetInstallReadme): Promise<TReadme> {
  const data = await fetch(
    `${API_URL}/v1/installs/${installId}/readme`,
    await getFetchOpts(orgId)
  )

  if (!data.ok) {
    throw new Error('Failed to fetch install readme')
  }

  return data.json()
}
