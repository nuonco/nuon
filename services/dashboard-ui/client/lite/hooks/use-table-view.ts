import { useCallback, useSyncExternalStore } from 'react'

export type TTableView = 'table' | 'cards'

export const TABLE_VIEW_STORAGE_KEY = 'nuon-lite-table-view'

const listeners = new Set<() => void>()
let memoryView: TTableView = 'table'

const isTableView = (value: string | null): value is TTableView =>
  value === 'table' || value === 'cards'

const getSnapshot = (): TTableView => {
  try {
    const stored = window.sessionStorage.getItem(TABLE_VIEW_STORAGE_KEY)
    memoryView = isTableView(stored) ? stored : 'table'
    return memoryView
  } catch {
    return memoryView
  }
}

const getServerSnapshot = (): TTableView => 'table'

const subscribe = (listener: () => void) => {
  const handleStorage = (event: StorageEvent) => {
    if (event.storageArea === window.sessionStorage) listener()
  }

  listeners.add(listener)
  window.addEventListener('storage', handleStorage)

  return () => {
    listeners.delete(listener)
    window.removeEventListener('storage', handleStorage)
  }
}

const writeView = (view: TTableView) => {
  memoryView = view
  try {
    window.sessionStorage.setItem(TABLE_VIEW_STORAGE_KEY, view)
  } catch {
    // The in-memory preference keeps the control usable when storage is blocked.
  }
  listeners.forEach((listener) => listener())
}

export const useTableView = () => {
  const view = useSyncExternalStore(subscribe, getSnapshot, getServerSnapshot)
  const setView = useCallback((next: TTableView) => writeView(next), [])

  return { view, setView }
}
