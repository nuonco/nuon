import type { TGroupStepSelection } from '@/components/installs/CreateInstall/GroupStep'
import type { TAppBranchConfig } from '@/types'

interface IStoredGroupSelection {
  branchId: string
  mode: TGroupStepSelection['mode']
  groupName: string | null
}

export const groupDraftKey = (storageKey: string) => `${storageKey}:group`

export const saveGroupDraft = (
  storageKey: string,
  branchId: string,
  selection: TGroupStepSelection
) => {
  try {
    const stored: IStoredGroupSelection = {
      branchId,
      mode: selection.mode,
      groupName: selection.group?.name ?? null,
    }
    localStorage.setItem(groupDraftKey(storageKey), JSON.stringify(stored))
  } catch (error) {
    console.warn('Failed to save group selection draft:', error)
  }
}

export const clearGroupDraft = (storageKey: string) => {
  try {
    localStorage.removeItem(groupDraftKey(storageKey))
  } catch (error) {
    console.warn('Failed to clear group selection draft:', error)
  }
}

export const resolveGroupDraft = (
  raw: string | null,
  branchId: string,
  config?: TAppBranchConfig
): TGroupStepSelection | null => {
  if (!raw) return null
  try {
    const stored = JSON.parse(raw) as IStoredGroupSelection
    if (stored.branchId !== branchId || stored.mode === 'none') return null
    const group = config?.install_groups?.find(
      (item) => item.name === stored.groupName
    )
    if (!group) return null
    return { mode: stored.mode, group }
  } catch {
    return null
  }
}

export const loadGroupDraft = (
  storageKey: string,
  branchId: string,
  config?: TAppBranchConfig
): TGroupStepSelection | null => {
  try {
    return resolveGroupDraft(
      localStorage.getItem(groupDraftKey(storageKey)),
      branchId,
      config
    )
  } catch {
    return null
  }
}
