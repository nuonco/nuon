export const LAST_APP_BRANCH_STORAGE_KEY = 'nuon-lite:last-app-branch'

type TLastAppBranchStore = Record<string, Record<string, string>>

const readStore = (): TLastAppBranchStore => {
  try {
    const raw = localStorage.getItem(LAST_APP_BRANCH_STORAGE_KEY)
    if (!raw) return {}
    const parsed = JSON.parse(raw) as unknown
    if (!parsed || typeof parsed !== 'object') return {}
    return parsed as TLastAppBranchStore
  } catch {
    return {}
  }
}

const writeStore = (store: TLastAppBranchStore) => {
  try {
    localStorage.setItem(LAST_APP_BRANCH_STORAGE_KEY, JSON.stringify(store))
  } catch {
    return
  }
}

export const getLastAppBranch = (orgId: string, appId: string) => {
  const branchId = readStore()[orgId]?.[appId]
  return branchId || undefined
}

export const setLastAppBranch = (
  orgId: string,
  appId: string,
  branchId: string
) => {
  const store = readStore()
  writeStore({
    ...store,
    [orgId]: {
      ...store[orgId],
      [appId]: branchId,
    },
  })
}

export const appSetupHref = (orgId: string, appId?: string) => {
  const path = `/${orgId}/apps/setup`
  if (!appId) return path
  return `${path}?${new URLSearchParams({ appId })}`
}

export const appBranchHref = (orgId: string, appId: string, branchId: string) =>
  `/${orgId}/apps/${appId}/branches/${branchId}`

export const resolveAppHref = ({
  orgId,
  appId,
  branchIds,
  lastBranchId,
}: {
  orgId: string
  appId: string
  branchIds: string[]
  lastBranchId?: string
}) => {
  if (branchIds.length === 0) return appSetupHref(orgId, appId)
  if (branchIds.length === 1) {
    return appBranchHref(orgId, appId, branchIds[0]!)
  }
  if (lastBranchId && branchIds.includes(lastBranchId)) {
    return appBranchHref(orgId, appId, lastBranchId)
  }
  return appBranchHref(orgId, appId, branchIds[0]!)
}
