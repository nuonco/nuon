import { useState } from 'react'
import type { VisibilityState } from '@tanstack/react-table'
import {
  columnVisibilityOverrides,
  mergeColumnVisibility,
  parseTableSettings,
  serializeTableSettings,
  tableSettingsKey,
  type ITableSettings,
  type TTableView,
} from '@/utils/table-settings-utils'

const readSettings = (tableId: string): ITableSettings => {
  try {
    return parseTableSettings(localStorage.getItem(tableSettingsKey(tableId)))
  } catch {
    return {}
  }
}

const writeSettings = (tableId: string, settings: ITableSettings) => {
  try {
    const value = serializeTableSettings(settings)
    if (value === null) localStorage.removeItem(tableSettingsKey(tableId))
    else localStorage.setItem(tableSettingsKey(tableId), value)
  } catch {}
}

export function useTableSettings(
  tableId: string,
  defaultVisibility: VisibilityState
) {
  const [settings, setSettings] = useState<ITableSettings>(() =>
    readSettings(tableId)
  )

  const update = (next: ITableSettings) => {
    setSettings(next)
    writeSettings(tableId, next)
  }

  const columnVisibility = mergeColumnVisibility(
    defaultVisibility,
    settings.columnVisibility
  )

  return {
    columnVisibility,
    isCustomized: Boolean(settings.columnVisibility || settings.view),
    reset: () => update({}),
    setColumnVisibility: (visibility: VisibilityState) =>
      update({
        ...settings,
        columnVisibility: columnVisibilityOverrides(
          defaultVisibility,
          visibility
        ),
      }),
    setView: (view: TTableView) => update({ ...settings, view }),
    view: settings.view,
  }
}
