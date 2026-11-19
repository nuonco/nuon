import { describe, expect, test } from 'bun:test'
import {
  columnVisibilityOverrides,
  mergeColumnVisibility,
  parseTableSettings,
  serializeTableSettings,
  tableSettingsKey,
} from './table-settings-utils'

describe('table-settings-utils', () => {
  test('tableSettingsKey namespaces the table id', () => {
    expect(tableSettingsKey('installs')).toBe('nuon-table:installs')
  })

  describe('parseTableSettings', () => {
    test('returns empty settings for missing or malformed values', () => {
      expect(parseTableSettings(null)).toEqual({})
      expect(parseTableSettings('')).toEqual({})
      expect(parseTableSettings('not json')).toEqual({})
      expect(parseTableSettings('[]')).toEqual({})
      expect(parseTableSettings('"grid"')).toEqual({})
    })

    test('keeps a valid view and boolean column overrides', () => {
      expect(
        parseTableSettings(
          JSON.stringify({
            view: 'grid',
            columnVisibility: { region: false, labels: true },
          })
        )
      ).toEqual({ view: 'grid', columnVisibility: { region: false, labels: true } })
    })

    test('drops an unknown view and non-boolean overrides', () => {
      expect(
        parseTableSettings(
          JSON.stringify({
            view: 'list',
            columnVisibility: { region: 'no', labels: true, group: null },
          })
        )
      ).toEqual({ columnVisibility: { labels: true } })
      expect(
        parseTableSettings(JSON.stringify({ columnVisibility: ['region'] }))
      ).toEqual({})
    })
  })

  describe('serializeTableSettings', () => {
    test('returns null when nothing is customised', () => {
      expect(serializeTableSettings({})).toBeNull()
    })

    test('round-trips through parseTableSettings', () => {
      const settings = { view: 'table' as const, columnVisibility: { region: false } }
      expect(parseTableSettings(serializeTableSettings(settings))).toEqual(settings)
    })
  })

  describe('mergeColumnVisibility', () => {
    test('applies overrides on top of defaults', () => {
      expect(
        mergeColumnVisibility(
          { name: true, region: true, labels: false },
          { region: false }
        )
      ).toEqual({ name: true, region: false, labels: false })
    })

    test('shows a column added after the overrides were saved', () => {
      expect(
        mergeColumnVisibility({ name: true, group: true }, { region: false })
      ).toEqual({ name: true, group: true, region: false })
    })
  })

  describe('columnVisibilityOverrides', () => {
    const defaults = { name: true, region: true, labels: false }

    test('stores only columns that differ from their default', () => {
      expect(
        columnVisibilityOverrides(defaults, {
          name: true,
          region: false,
          labels: true,
        })
      ).toEqual({ region: false, labels: true })
    })

    test('returns undefined when everything matches the defaults', () => {
      expect(columnVisibilityOverrides(defaults, defaults)).toBeUndefined()
    })

    test('treats columns without a default as visible', () => {
      expect(columnVisibilityOverrides({}, { extra: true })).toBeUndefined()
      expect(columnVisibilityOverrides({}, { extra: false })).toEqual({
        extra: false,
      })
    })
  })
})
