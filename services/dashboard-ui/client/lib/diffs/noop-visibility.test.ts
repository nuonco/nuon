import { describe, expect, test } from 'bun:test'
import { terraformActionSelected } from './noop-visibility'

const selected = new Set(['create', 'update', 'delete', 'replace'])

describe('terraformActionSelected', () => {
  test('hides no-ops until the toggle is on', () => {
    expect(terraformActionSelected('no-op', selected, false)).toBe(false)
    expect(terraformActionSelected('no-op', selected, true)).toBe(true)
  })

  test('leaves other actions on the filter selection', () => {
    expect(terraformActionSelected('create', selected, false)).toBe(true)
    expect(terraformActionSelected('read', selected, true)).toBe(false)
  })
})
