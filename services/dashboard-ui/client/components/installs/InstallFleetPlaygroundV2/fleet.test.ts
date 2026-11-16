import { describe, expect, test } from 'bun:test'
import { mixedFleet, mixedFleetConfigs } from './fixtures'
import {
  groupFleet,
  labelOptions,
  matchesFilters,
  summarizeFleet,
  type TFleetFilters,
} from './fleet'

const noFilters: TFleetFilters = { search: '', labels: new Set() }

describe('groupFleet', () => {
  test('orders branches by name and puts installs with no branch last', () => {
    expect(groupFleet(mixedFleet).map((branch) => branch.name ?? null)).toEqual(
      ['main', 'monthly', null]
    )
  })

  test('groups installs by install group, worst health first', () => {
    const main = groupFleet(mixedFleet)[0]
    expect(main.groups.map((group) => group.name)).toEqual([
      'production',
      'stage',
    ])
    expect(main.groups[0].installs.map((install) => install.name)).toEqual([
      'acme-prod-ap',
      'acme-prod-eu',
      'acme-prod-us',
    ])
    expect(main.groups[0].worst).toBe('unhealthy')
  })

  test('attaches the branch config label selector to the group', () => {
    const branches = groupFleet(mixedFleet, mixedFleetConfigs)
    expect(
      branches[0].groups.find((group) => group.name === 'production')?.labels
    ).toEqual({
      env: 'prod',
    })
    expect(branches[1].groups[0].labels).toEqual({
      cadence: 'monthly',
      env: 'prod',
    })
  })

  test('orders install groups by their rollout order in the branch config', () => {
    const main = groupFleet(mixedFleet, mixedFleetConfigs)[0]
    expect(main.groups.map((group) => group.name)).toEqual([
      'stage',
      'production',
    ])
  })
})

describe('summarizeFleet', () => {
  test('buckets health like GET /v1/installs/health', () => {
    expect(summarizeFleet(mixedFleet)).toEqual({
      total: 7,
      healthy: 3,
      degraded: 1,
      unhealthy: 1,
      unknown: 1,
      noSignal: 1,
      runnerNotActive: 2,
    })
  })
})

describe('matchesFilters', () => {
  const names = (filters: Partial<TFleetFilters>) =>
    mixedFleet
      .filter((install) => matchesFilters(install, { ...noFilters, ...filters }))
      .map((install) => install.name)

  test('requires every selected label', () => {
    expect(names({ labels: new Set(['env:prod', 'region:eu-west-1']) })).toEqual(
      ['acme-prod-eu']
    )
  })

  test('filters by any selected health, keeping progressing distinct from unknown', () => {
    expect(names({ health: new Set(['progressing']) })).toEqual(['acme-stage-eu'])
    expect(names({ health: new Set(['no-signal']) })).toEqual(['acme-labs'])
    expect(names({ health: new Set(['degraded', 'unhealthy']) })).toEqual([
      'acme-prod-eu',
      'acme-prod-ap',
    ])
  })

  test('filters by runner state', () => {
    expect(names({ runner: 'not-active' })).toEqual([
      'acme-prod-ap',
      'acme-labs',
    ])
  })

  test('searches name and ID', () => {
    expect(names({ search: 'inst-bank' })).toEqual(['acme-bank'])
  })
})

test('labelOptions lists distinct key:value pairs', () => {
  expect(labelOptions(mixedFleet)).toContain('tier:canary')
  expect(new Set(labelOptions(mixedFleet)).size).toBe(
    labelOptions(mixedFleet).length
  )
})
