import { describe, expect, test } from 'bun:test'
import {
  mixedFleet,
  mixedFleetBranches,
  mixedFleetConfigs,
  mixedFleetOverviews,
} from './fixtures'
import {
  componentOptions,
  groupFleet,
  labelOptions,
  latestConfigIdByBranch,
  matchesFilters,
  rolloutState,
  summarizeFleet,
  type TFleetContext,
  type TFleetFilters,
} from './fleet'

const noFilters: TFleetFilters = { search: '', labels: new Set() }

const context: TFleetContext = {
  latestConfigIdByBranch: latestConfigIdByBranch(mixedFleetBranches),
  overviewById: mixedFleetOverviews,
}

describe('groupFleet', () => {
  test('orders branches by name and puts installs with no branch last', () => {
    expect(groupFleet(mixedFleet).map((branch) => branch.name ?? null)).toEqual(
      ['main', 'monthly', null]
    )
  })

  test('puts installs behind the latest run first, then worst health', () => {
    const main = groupFleet(mixedFleet, [], context)[0]
    expect(main.groups.map((group) => group.name)).toEqual([
      'production',
      'stage',
    ])
    expect(main.groups[0].installs.map((install) => install.name)).toEqual([
      'acme-prod-ap',
      'acme-prod-eu',
      'acme-prod-us',
    ])
  })

  test('attaches the branch config label selector to the group', () => {
    const branches = groupFleet(mixedFleet, mixedFleetConfigs, context)
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
    const main = groupFleet(mixedFleet, mixedFleetConfigs, context)[0]
    expect(main.groups.map((group) => group.name)).toEqual([
      'stage',
      'production',
    ])
  })
})

describe('rolloutState', () => {
  const state = (name: string) =>
    rolloutState(
      mixedFleet.find((install) => install.name === name)!,
      context
    )

  test('compares the install config to the latest branch run config', () => {
    expect(state('acme-prod-us')).toBe('applied')
    expect(state('acme-prod-eu')).toBe('behind')
  })

  test('is unknown without a branch or without a latest run', () => {
    expect(state('acme-labs')).toBe('unknown')
    expect(
      rolloutState(mixedFleet[0], { latestConfigIdByBranch: {} })
    ).toBe('unknown')
  })
})

describe('summarizeFleet', () => {
  test('counts installs on and behind the latest run', () => {
    expect(summarizeFleet(mixedFleet, context)).toEqual({
      total: 7,
      applied: 4,
      behind: 2,
      runnerNotActive: 2,
    })
  })
})

describe('matchesFilters', () => {
  const names = (filters: Partial<TFleetFilters>) =>
    mixedFleet
      .filter((install) =>
        matchesFilters(install, { ...noFilters, ...filters }, context)
      )
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

  test('filters by whether the latest branch run is applied', () => {
    expect(names({ rollout: 'behind' })).toEqual([
      'acme-prod-eu',
      'acme-prod-ap',
    ])
    expect(names({ rollout: 'applied' })).toEqual([
      'acme-canary',
      'acme-stage-eu',
      'acme-prod-us',
      'acme-bank',
    ])
  })

  test('filters by whether a component is live', () => {
    expect(names({ component: 'comp-worker' })).toEqual([
      'acme-canary',
      'acme-stage-eu',
      'acme-prod-us',
      'acme-bank',
    ])
    expect(
      names({ component: 'comp-worker', componentState: 'not-live' })
    ).toEqual(['acme-prod-eu', 'acme-prod-ap'])
  })

  test('treats a component the install has no drift data for as neither live nor not live', () => {
    expect(names({ component: 'comp-web', componentState: 'not-live' })).toEqual(
      []
    )
    expect(names({ component: 'comp-web' })).not.toContain('acme-labs')
  })
})

test('componentOptions lists each component once, by name', () => {
  expect(componentOptions(mixedFleetOverviews)).toEqual([
    { value: 'comp-api', label: 'api' },
    { value: 'comp-web', label: 'web' },
    { value: 'comp-worker', label: 'worker' },
  ])
})

test('labelOptions lists distinct key:value pairs', () => {
  expect(labelOptions(mixedFleet)).toContain('tier:canary')
  expect(new Set(labelOptions(mixedFleet)).size).toBe(
    labelOptions(mixedFleet).length
  )
})
