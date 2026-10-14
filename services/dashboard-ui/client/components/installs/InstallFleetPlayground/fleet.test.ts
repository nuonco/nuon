import { describe, expect, test } from 'bun:test'
import {
  groupFleetInstalls,
  labelOptions,
  layoutFleet,
  matchesSearch,
  passesFilters,
  type TFleetInstall,
} from './fleet'

const install = (
  overrides: Partial<TFleetInstall> & Pick<TFleetInstall, 'id' | 'name'>
): TFleetInstall => ({
  labels: {},
  lifecycle: 'provisioned',
  hasRunner: true,
  runnerStatus: 'active',
  configDriftCount: 0,
  commitsBehind: 0,
  ...overrides,
})

describe('fleet grouping', () => {
  test('puts the current branch first, other branches next, and untracked last', () => {
    const sections = groupFleetInstalls(
      [
        install({ id: '1', name: 'zeta', branchName: 'staging' }),
        install({ id: '2', name: 'alpha' }),
        install({ id: '3', name: 'beta', branchName: 'main' }),
        install({ id: '4', name: 'gamma', branchName: 'feat/multi-region' }),
      ],
      'main'
    )
    expect(sections.map((section) => section.title)).toEqual([
      'main',
      'feat/multi-region',
      'staging',
      'Not tracking',
    ])
  })

  test('sorts installs in a branch by name', () => {
    const [section] = groupFleetInstalls(
      [
        install({ id: 'b', name: 'beta', branchName: 'main' }),
        install({ id: 'a', name: 'alpha', branchName: 'main' }),
      ],
      'main'
    )
    expect(section.installs.map((item) => item.name)).toEqual(['alpha', 'beta'])
  })

  test('wraps sections after four columns', () => {
    const layout = layoutFleet(
      ['a', 'b', 'c', 'd', 'e'].map((title) => ({
        id: title,
        title,
        installs: [install({ id: title, name: title, branchName: title })],
      }))
    )
    expect(
      layout.map((section) => [section.title, section.x, section.y > 0])
    ).toEqual([
      ['a', 0, false],
      ['b', 288, false],
      ['c', 576, false],
      ['d', 864, false],
      ['e', 0, true],
    ])
  })
})

describe('fleet filters', () => {
  const none = {
    labels: new Set<string>(),
    needsUpdate: false,
  }
  const prod = install({
    id: 'prod',
    name: 'acme-prod',
    branchName: 'main',
    labels: { env: 'prod', region: 'us-east-1' },
    commitsBehind: 2,
  })
  const quiet = install({
    id: 'quiet',
    name: 'acme-labs',
    hasRunner: false,
    runnerStatus: undefined,
    commitsBehind: undefined,
    labels: { env: 'prod' },
  })

  test('requires every selected label', () => {
    const labels = new Set(['env:prod', 'region:us-east-1'])
    expect(passesFilters(prod, { ...none, labels })).toBe(true)
    expect(passesFilters(quiet, { ...none, labels })).toBe(false)
  })

  test('treats a missing runner as not active', () => {
    expect(passesFilters(quiet, { ...none, runner: 'inactive' })).toBe(true)
    expect(passesFilters(prod, { ...none, runner: 'active' })).toBe(true)
    expect(passesFilters(quiet, { ...none, runner: 'active' })).toBe(false)
    expect(passesFilters(quiet, none)).toBe(true)
    expect(passesFilters(prod, none)).toBe(true)
  })

  test('needs updating ignores current and unknown tracking', () => {
    expect(passesFilters(prod, { ...none, needsUpdate: true })).toBe(true)
    expect(passesFilters(quiet, { ...none, needsUpdate: true })).toBe(false)
    expect(
      passesFilters(
        install({ id: 'current', name: 'current', commitsBehind: 0 }),
        { ...none, needsUpdate: true }
      )
    ).toBe(false)
  })

  test('search matches name, label, and runner', () => {
    expect(matchesSearch(prod, 'us-east-1')).toBe(true)
    expect(matchesSearch(quiet, 'no runner')).toBe(true)
    expect(matchesSearch(prod, 'staging')).toBe(false)
  })

  test('collects label options from the fleet', () => {
    expect(labelOptions([prod, quiet])).toEqual([
      'env:prod',
      'region:us-east-1',
    ])
  })
})
