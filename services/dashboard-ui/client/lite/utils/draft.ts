export const DRAFT_STORAGE_PREFIX = 'nuon-lite:draft'

export const draftStorageKey = (
  orgId: string,
  wizard: string,
  resourceId?: string
) => `${DRAFT_STORAGE_PREFIX}:${orgId}:${wizard}:${resourceId ?? 'new'}`

export interface IDraftRecord<T> {
  values: T
  updatedAt: string
}

const parseRecord = <T>(raw: string | null): IDraftRecord<T> | undefined => {
  if (!raw) return undefined
  try {
    const parsed = JSON.parse(raw) as unknown
    if (!parsed || typeof parsed !== 'object') return undefined
    const record = parsed as Partial<IDraftRecord<T>>
    if (record.values === undefined || typeof record.updatedAt !== 'string') {
      return undefined
    }
    return { values: record.values, updatedAt: record.updatedAt }
  } catch {
    return undefined
  }
}

const readDraft = <T>(key: string) => {
  try {
    return parseRecord<T>(localStorage.getItem(key))
  } catch {
    return undefined
  }
}

const writeDraft = <T>(key: string, record: IDraftRecord<T>) => {
  try {
    localStorage.setItem(key, JSON.stringify(record))
  } catch {
    return
  }
}

const removeDraft = (key: string) => {
  try {
    localStorage.removeItem(key)
  } catch {
    return
  }
}

export const loadDraft = <T>(
  orgId: string,
  wizard: string,
  resourceId?: string
) => readDraft<T>(draftStorageKey(orgId, wizard, resourceId))

export const saveDraft = <T>(
  orgId: string,
  wizard: string,
  values: T,
  resourceId?: string
) => {
  writeDraft(draftStorageKey(orgId, wizard, resourceId), {
    values,
    updatedAt: new Date().toISOString(),
  })
}

export const clearDraft = (
  orgId: string,
  wizard: string,
  resourceId?: string
) => {
  removeDraft(draftStorageKey(orgId, wizard, resourceId))
}

export const migrateDraft = (orgId: string, wizard: string, resourceId: string) => {
  const created = readDraft(draftStorageKey(orgId, wizard))
  const existing = readDraft(draftStorageKey(orgId, wizard, resourceId))
  if (created && !existing) {
    writeDraft(draftStorageKey(orgId, wizard, resourceId), created)
  }
  removeDraft(draftStorageKey(orgId, wizard))
}
