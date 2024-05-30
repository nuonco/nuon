import type { TInstallComponent } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export interface IGetInstallComponent {
  installComponentId: string
  installId: string
  orgId: string
}

export async function getInstallComponent({
  installComponentId,
  installId,
  orgId,
}: IGetInstallComponent): Promise<TInstallComponent> {
  const res = await fetch(
    `${API_URL}/v1/installs/${installId}/components`,
    await getFetchOpts(orgId)
  )

  if (!res.ok) {
    throw new Error('Failed to fetch install component')
  }

  return res
    .json()
    .then((comps: Array<TInstallComponent>) =>
      comps.find((c) => c?.id === installComponentId)
    )
}
