import { getAppInstalls } from './get-app-installs'

export interface IInstallNameTaken {
  appId: string
  orgId: string
  name: string
}

const PAGE_SIZE = 100

// The `q` param is a substring search, so every page has to be compared for an
// exact match before the name can be treated as available.
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
