import { describe, expect, test } from 'bun:test'
import {
  mixedFleet,
  mixedFleetApprovals,
  mixedFleetBranches,
  mixedFleetConfigVersions,
  mixedFleetConfigs,
  mixedFleetOverviews,
  mixedFleetRuns,
} from './fixtures'
import {
  approvalsByInstall,
  attentionItems,
  disabledRolesText,
  installOverrides,
  overridesText,
  componentOptions,
  configVersion,
  countOnRun,
  groupFleet,
  labelOptions,
  latestConfigIdByBranch,
  matchesFilters,
  rolloutState,
  runsIndex,
  summarizeFleet,
  type TFleetContext,
  type TFleetFilters,
} from './fleet'

const noFilters: TFleetFilters = { search: '', labels: new Set() }

const context: TFleetContext = {
  latestConfigIdByBranch: latestConfigIdByBranch(mixedFleetBranches),
  runsById: runsIndex(mixedFleetRuns),
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
    rolloutState(mixedFleet.find((install) => install.name === name)!, context)

  test('compares the install config to the latest branch run config', () => {
    expect(state('acme-prod-us')).toBe('applied')
    expect(state('acme-prod-eu')).toBe('behind')
  })

  test('is unknown without a branch or without a latest run', () => {
    expect(state('acme-labs')).toBe('unknown')
    expect(rolloutState(mixedFleet[0], { latestConfigIdByBranch: {} })).toBe(
      'unknown'
    )
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
  const names = (
    filters: Partial<TFleetFilters>,
    extra: Partial<TFleetContext> = {}
  ) =>
    mixedFleet
      .filter((install) =>
        matchesFilters(
          install,
          { ...noFilters, ...filters },
          { ...context, ...extra }
        )
      )
      .map((install) => install.name)

  test('requires every selected label', () => {
    expect(
      names({ labels: new Set(['env:prod', 'region:eu-west-1']) })
    ).toEqual(['acme-prod-eu'])
  })

  test('filters by any selected health, keeping progressing distinct from unknown', () => {
    expect(names({ health: new Set(['progressing']) })).toEqual([
      'acme-stage-eu',
    ])
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

  test('filters one branch to installs on or not on its run, leaving other branches alone', () => {
    expect(names({ rolloutByBranch: { 'br-main': 'behind' } })).toEqual([
      'acme-prod-eu',
      'acme-prod-ap',
      'acme-bank',
      'acme-labs',
    ])
  })

  test('compares against the run picked for that branch', () => {
    expect(
      names(
        { rolloutByBranch: { 'br-main': 'applied' } },
        { runByBranch: { 'br-main': 'run-main-2' } }
      )
    ).toEqual(['acme-prod-eu', 'acme-prod-ap', 'acme-bank', 'acme-labs'])
  })

  test('compares one component when a scope is set', () => {
    expect(
      names(
        { rolloutByBranch: { 'br-main': 'applied' } },
        { scope: 'comp-api' }
      )
    ).toEqual(['acme-canary', 'acme-stage-eu', 'acme-bank', 'acme-labs'])
  })
})

describe('rolloutState with a picked run', () => {
  const state = (name: string, extra: Partial<TFleetContext>) =>
    rolloutState(mixedFleet.find((install) => install.name === name)!, {
      ...context,
      ...extra,
    })

  test('installs on an older run read as applied for that run', () => {
    expect(
      state('acme-prod-eu', { runByBranch: { 'br-main': 'run-main-2' } })
    ).toBe('applied')
    expect(
      state('acme-prod-us', { runByBranch: { 'br-main': 'run-main-2' } })
    ).toBe('behind')
  })

  test('a lagging component reads as behind even when the install took the run', () => {
    expect(state('acme-prod-us', { scope: 'comp-api' })).toBe('behind')
    expect(state('acme-prod-us', {})).toBe('applied')
  })

  test('the summary ignores the picked run and scope', () => {
    expect(
      summarizeFleet(mixedFleet, {
        ...context,
        runByBranch: { 'br-main': 'run-main-2' },
        scope: 'comp-api',
      })
    ).toEqual({ total: 7, applied: 4, behind: 2, runnerNotActive: 2 })
  })
})

test('attentionItems merges failing health and pending approvals, worst first', () => {
  expect(
    attentionItems(mixedFleet, approvalsByInstall(mixedFleetApprovals)).map(
      (item) => [item.install.name, item.reasons]
    )
  ).toEqual([
    ['acme-prod-ap', ['unhealthy']],
    ['acme-canary', ['pending-approval']],
    ['acme-prod-eu', ['pending-approval', 'degraded']],
  ])
})

test('approvalsByInstall ignores approvals not owned by an install', () => {
  expect(
    approvalsByInstall([
      { id: 'apr-run', owner_type: 'app_branches', owner_id: 'br-main' },
    ])
  ).toEqual({})
})

test('disabledRolesText counts roles that are turned off', () => {
  const byName = (name: string) =>
    mixedFleet.find((install) => install.name === name)!
  expect(disabledRolesText(byName('acme-prod-ap'))).toBe('2 roles disabled')
  expect(disabledRolesText(byName('acme-prod-eu'))).toBe('1 role disabled')
  expect(disabledRolesText(byName('acme-canary'))).toBeUndefined()
})

test('overridesText counts components with an override value set', () => {
  const byName = (name: string) =>
    mixedFleet.find((install) => install.name === name)!
  expect(overridesText(byName('acme-canary'))).toBe('2 overrides')
  expect(installOverrides(byName('acme-canary'))).toEqual(['api', 'web'])
  expect(overridesText(byName('acme-prod-us'))).toBe('1 override')
  expect(overridesText(byName('acme-prod-eu'))).toBeUndefined()
})

test('configVersion maps the install config to its version', () => {
  expect(
    configVersion(mixedFleet[0], {
      configVersionById: mixedFleetConfigVersions,
    })
  ).toBe(3)
})

test('countOnRun counts installs on a run', () => {
  expect(countOnRun(mixedFleet, 'run-main-3', undefined, context)).toBe(3)
  expect(countOnRun(mixedFleet, 'run-main-3', 'comp-api', context)).toBe(2)
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
