import { API_URL, getFetchOpts } from '@/utils'

// TODO: remove this once type generation is fixed.
// Is the new type in the swagger spec?
type TInstallReadme = {
  readme: string
}

export interface IGetInstallReadme {
  installId: string
  orgId: string
}

export async function getInstallReadme({
  installId,
  orgId,
}: IGetInstallReadme): Promise<TInstallReadme> {
  const data = await fetch(
    `${API_URL}/v1/installs/${installId}/readme`,
    await getFetchOpts(orgId)
  )

  if (!data.ok) {
    throw new Error('Failed to fetch install readme')
  }

  return data.json()
}
