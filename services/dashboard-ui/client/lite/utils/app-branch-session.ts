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
