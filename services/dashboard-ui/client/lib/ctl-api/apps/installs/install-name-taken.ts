import { getAppInstalls } from './get-app-installs'

export interface IInstallNameTaken {
  appId: string
  orgId: string
  name: string
}

const PAGE_SIZE = 100

export async function installNameTaken({
  appId,
  orgId,
  name,
}: IInstallNameTaken): Promise<boolean> {
  const trimmed = name.trim()
  if (!trimmed) return false

  let offset = 0
  for (;;) {
    const { data, pagination } = await getAppInstalls({
      appId,
      orgId,
      q: trimmed,
      limit: PAGE_SIZE,
      offset,
    })

    const installs = data ?? []
    if (installs.some((install) => install?.name === trimmed)) return true
    if (!pagination?.hasNext || installs.length === 0) return false

    offset += installs.length
  }
}
