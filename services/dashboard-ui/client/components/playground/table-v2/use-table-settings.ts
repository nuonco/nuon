import { useCallback, useState } from 'react'
import type { VisibilityState } from '@tanstack/react-table'

export type TTableV2View = 'table' | 'grid'

interface ITableSettings {
  columnVisibility?: VisibilityState
  view?: TTableV2View
}

const storageKey = (tableId: string) => `nuon-table:${tableId}`

const readSettings = (tableId: string): ITableSettings => {
  try {
    const raw = localStorage.getItem(storageKey(tableId))
    if (!raw) return {}
    const parsed = JSON.parse(raw)
    return parsed && typeof parsed === 'object' ? parsed : {}
  } catch {
    return {}
  }
}

const writeSettings = (tableId: string, settings: ITableSettings) => {
  try {
    if (!settings.columnVisibility && !settings.view) {
      localStorage.removeItem(storageKey(tableId))
    } else {
      localStorage.setItem(storageKey(tableId), JSON.stringify(settings))
    }
  } catch {}
}

export function useTableSettings(
  tableId: string,
  defaultVisibility: VisibilityState
) {
  const [settings, setSettings] = useState<ITableSettings>(() =>
    readSettings(tableId)
  )

  const update = useCallback(
    (next: ITableSettings) => {
      setSettings(next)
      writeSettings(tableId, next)
    },
    [tableId]
  )

  const columnVisibility: VisibilityState = {
    ...defaultVisibility,
    ...settings.columnVisibility,
  }

  const setColumnVisibility = (visibility: VisibilityState) => {
    const overrides = Object.fromEntries(
      Object.entries(visibility).filter(
        ([id, visible]) => (defaultVisibility[id] ?? true) !== visible
      )
    )
    update({
      ...settings,
      columnVisibility: Object.keys(overrides).length ? overrides : undefined,
    })
  }

  const setView = (view: TTableV2View) => update({ ...settings, view })

  const reset = () => update({})

  return {
    columnVisibility,
    hasCustomColumns: Boolean(settings.columnVisibility),
    reset,
    setColumnVisibility,
    setView,
    view: settings.view,
  }
}
