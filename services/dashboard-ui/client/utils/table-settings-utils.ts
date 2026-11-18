import type { VisibilityState } from '@tanstack/react-table'

export type TTableView = 'table' | 'grid'

export const TABLE_VIEWS: readonly TTableView[] = ['table', 'grid']

export interface ITableSettings {
  columnVisibility?: VisibilityState
  view?: TTableView
}

export const tableSettingsKey = (tableId: string) => `nuon-table:${tableId}`

export const parseTableSettings = (raw: string | null): ITableSettings => {
  if (!raw) return {}
  let parsed: unknown
  try {
    parsed = JSON.parse(raw)
  } catch {
    return {}
  }
  if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) return {}

  const { columnVisibility, view } = parsed as Record<string, unknown>
  const settings: ITableSettings = {}

  if (typeof view === 'string' && (TABLE_VIEWS as string[]).includes(view)) {
    settings.view = view as TTableView
  }

  if (
    columnVisibility &&
    typeof columnVisibility === 'object' &&
    !Array.isArray(columnVisibility)
  ) {
    const entries = Object.entries(columnVisibility).filter(
      ([, visible]) => typeof visible === 'boolean'
    )
    if (entries.length) settings.columnVisibility = Object.fromEntries(entries)
  }

  return settings
}

export const serializeTableSettings = (
  settings: ITableSettings
): string | null =>
  settings.columnVisibility || settings.view ? JSON.stringify(settings) : null

export const mergeColumnVisibility = (
  defaults: VisibilityState,
  overrides?: VisibilityState
): VisibilityState => ({ ...defaults, ...overrides })

export const columnVisibilityOverrides = (
  defaults: VisibilityState,
  visibility: VisibilityState
): VisibilityState | undefined => {
  const overrides = Object.fromEntries(
    Object.entries(visibility).filter(
      ([id, visible]) => (defaults[id] ?? true) !== visible
    )
  )
  return Object.keys(overrides).length ? overrides : undefined
}
