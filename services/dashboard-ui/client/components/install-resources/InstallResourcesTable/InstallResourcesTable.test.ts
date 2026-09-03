import { describe, expect, test } from 'bun:test'
import type { TInstallResource } from '@/types'
import { DateTime } from 'luxon'
import {
  groupComponentResources,
  hasHealthSignal,
  healthFacetCounts,
  matchesHealthFilter,
  matchesResourceSearch,
  NO_SIGNAL_FILTER,
  visibleRowCount,
} from './InstallResourcesTable'

const now = DateTime.now().toISO()
const longAgo = DateTime.now().minus({ hours: 2 }).toISO()

function resource(overrides: Partial<TInstallResource>): TInstallResource {
  return {
    install_component_id: 'instcmp-1',
    source: 'component',
    kind: 'Deployment',
    namespace: 'default',
    name: 'web',
    health: 'healthy',
    provider: 'kubernetes',
    observed_at: now,
    ...overrides,
  } as TInstallResource
}

describe('hasHealthSignal', () => {
  test('assessed kubernetes resources carry signal', () => {
    expect(hasHealthSignal(resource({ health: 'degraded' }))).toBe(true)
  })

  test('unknown and not-applicable carry none', () => {
    expect(hasHealthSignal(resource({ health: 'unknown' }))).toBe(false)
    expect(hasHealthSignal(resource({ health: 'not-applicable' }))).toBe(false)
  })

  test('cloud identity rows never carry signal, even if assessed', () => {
    expect(
      hasHealthSignal(resource({ provider: 'aws', health: 'healthy' }))
    ).toBe(false)
    expect(
      hasHealthSignal(resource({ provider: 'gcp', health: 'healthy' }))
    ).toBe(false)
    expect(
      hasHealthSignal(resource({ provider: 'azure', health: 'healthy' }))
    ).toBe(false)
  })

  test('rows removed from config carry none', () => {
    expect(hasHealthSignal(resource({ removed_from_config: true }))).toBe(false)
  })
})

describe('matchesHealthFilter', () => {
  test('an empty filter matches everything', () => {
    expect(matchesHealthFilter(resource({ health: 'unknown' }), '')).toBe(true)
  })

  test('an exact status filter matches on status', () => {
    expect(
      matchesHealthFilter(resource({ health: 'degraded' }), 'degraded')
    ).toBe(true)
    expect(
      matchesHealthFilter(resource({ health: 'healthy' }), 'degraded')
    ).toBe(false)
  })

  test('the no-signal filter matches every unassessed row', () => {
    expect(
      matchesHealthFilter(resource({ health: 'unknown' }), NO_SIGNAL_FILTER)
    ).toBe(true)
    expect(
      matchesHealthFilter(
        resource({ provider: 'aws', health: 'healthy' }),
        NO_SIGNAL_FILTER
      )
    ).toBe(true)
    expect(
      matchesHealthFilter(resource({ health: 'healthy' }), NO_SIGNAL_FILTER)
    ).toBe(false)
  })
})

describe('healthFacetCounts', () => {
  test('buckets every unassessed row under no-signal', () => {
    const counts = healthFacetCounts([
      resource({ health: 'healthy' }),
      resource({ health: 'healthy' }),
      resource({ health: 'unhealthy' }),
      resource({ health: 'unknown' }),
      resource({ health: 'not-applicable' }),
      resource({ provider: 'aws', health: 'healthy' }),
    ])
    expect(counts).toEqual({ healthy: 2, unhealthy: 1, [NO_SIGNAL_FILTER]: 3 })
  })
})

describe('groupComponentResources', () => {
  const groups = groupComponentResources(
    [
      resource({ install_component_id: 'a', name: 'a-ok', health: 'healthy' }),
      resource({
        install_component_id: 'a',
        name: 'a-cm',
        health: 'not-applicable',
      }),
      resource({
        install_component_id: 'b',
        name: 'b-bad',
        health: 'unhealthy',
      }),
      resource({ install_component_id: 'b', name: 'b-ok', health: 'healthy' }),
      resource({ install_component_id: 'c', name: 'c-iam', provider: 'aws' }),
    ],
    { a: 'alpha', b: 'bravo', c: 'charlie' }
  )

  test('sorts groups worst-health-first, not alphabetically', () => {
    expect(groups.map((g) => g.heading)).toEqual(['bravo', 'alpha', 'charlie'])
  })

  test('splits each group into signal and no-signal rows', () => {
    const alpha = groups.find((g) => g.heading === 'alpha')!
    expect(alpha.signalRows.map((r) => r.name)).toEqual(['a-ok'])
    expect(alpha.noSignalRows.map((r) => r.name)).toEqual(['a-cm'])
  })

  test('a group of only identity rows has no signal rows at all', () => {
    const charlie = groups.find((g) => g.heading === 'charlie')!
    expect(charlie.signalRows).toHaveLength(0)
    expect(charlie.noSignalRows).toHaveLength(1)
  })

  test('sorts rows within a group worst-first', () => {
    const bravo = groups.find((g) => g.heading === 'bravo')!
    expect(bravo.rows.map((r) => r.name)).toEqual(['b-bad', 'b-ok'])
  })

  test('counts failing live rows', () => {
    const bravo = groups.find((g) => g.heading === 'bravo')!
    expect(bravo.failing).toBe(1)
    expect(bravo.live).toBe(2)
    expect(bravo.worst).toBe('unhealthy')
  })
})

describe('matchesResourceSearch', () => {
  test('an empty search matches everything', () => {
    expect(matchesResourceSearch(resource({}), '')).toBe(true)
    expect(matchesResourceSearch(resource({}), '   ')).toBe(true)
  })

  test('matches name, kind, and namespace case-insensitively', () => {
    const r = resource({
      name: 'web-app',
      kind: 'Deployment',
      namespace: 'prod',
    })
    expect(matchesResourceSearch(r, 'WEB')).toBe(true)
    expect(matchesResourceSearch(r, 'deploy')).toBe(true)
    expect(matchesResourceSearch(r, 'PROD')).toBe(true)
    expect(matchesResourceSearch(r, 'postgres')).toBe(false)
  })
})

describe('group staleness', () => {
  test('a group is fully stale only when every reporting row is stale', () => {
    const [group] = groupComponentResources(
      [
        resource({
          install_component_id: 'a',
          name: 'one',
          observed_at: longAgo,
        }),
        resource({
          install_component_id: 'a',
          name: 'two',
          observed_at: longAgo,
        }),
      ],
      { a: 'alpha' }
    )
    expect(group.fullyStale).toBe(true)
    expect(group.lastReportedAt).toBe(longAgo)
  })

  test('one fresh row keeps the group reporting', () => {
    const [group] = groupComponentResources(
      [
        resource({
          install_component_id: 'a',
          name: 'one',
          observed_at: longAgo,
        }),
        resource({ install_component_id: 'a', name: 'two', observed_at: now }),
      ],
      { a: 'alpha' }
    )
    expect(group.fullyStale).toBe(false)
    expect(group.lastReportedAt).toBe(now)
  })

  test('identity and removed rows cannot make a group look stale', () => {
    const [group] = groupComponentResources(
      [
        resource({
          install_component_id: 'a',
          provider: 'aws',
          observed_at: longAgo,
        }),
        resource({
          install_component_id: 'a',
          removed_from_config: true,
          observed_at: longAgo,
        }),
      ],
      { a: 'alpha' }
    )
    expect(group.fullyStale).toBe(false)
  })
})

describe('group ordering with staleness', () => {
  test('a silent group outranks healthy and progressing, but not a live failure', () => {
    const groups = groupComponentResources(
      [
        resource({ install_component_id: 'a', health: 'healthy' }),
        resource({ install_component_id: 'b', health: 'progressing' }),
        resource({
          install_component_id: 'c',
          health: 'healthy',
          observed_at: longAgo,
        }),
        resource({ install_component_id: 'd', health: 'unhealthy' }),
      ],
      { a: 'alpha', b: 'bravo', c: 'silent', d: 'broken' }
    )
    expect(groups.map((g) => g.heading)).toEqual([
      'broken',
      'silent',
      'bravo',
      'alpha',
    ])
  })
})

describe('visibleRowCount', () => {
  function rowsFor(healths: string[]) {
    return groupComponentResources(
      healths.map((health, idx) =>
        resource({ install_component_id: 'a', name: `r-${idx}`, health })
      ),
      { a: 'alpha' }
    )[0].signalRows
  }

  test('shows every row when the group is small', () => {
    const rows = rowsFor(['healthy', 'healthy', 'degraded'])
    expect(rows.length).toBeLessThanOrEqual(visibleRowCount(rows))
  })

  test('folds only the healthy tail of a large group', () => {
    const rows = rowsFor([...Array(20).fill('healthy'), 'degraded'])
    const hidden = rows.slice(visibleRowCount(rows))
    expect(hidden.length).toBe(13)
    expect(hidden.every((row) => row.health === 'healthy')).toBe(true)
  })

  test('never folds a row that needs attention, however many there are', () => {
    const rows = rowsFor([
      ...Array(15).fill('unhealthy'),
      ...Array(15).fill('healthy'),
    ])
    const visible = rows.slice(0, visibleRowCount(rows))
    expect(visible.filter((row) => row.health === 'unhealthy')).toHaveLength(15)
    expect(
      rows.slice(visibleRowCount(rows)).every((r) => r.health === 'healthy')
    ).toBe(true)
  })

  test('progressing rows are never folded either', () => {
    const rows = rowsFor([
      ...Array(30).fill('healthy'),
      ...Array(4).fill('progressing'),
    ])
    const hidden = rows.slice(visibleRowCount(rows))
    expect(hidden.every((row) => row.health === 'healthy')).toBe(true)
  })
})

describe('group badge is the live roll-up', () => {
  // The debounced verdict is rendered by the 90-day card above this section;
  // mixing it into a heading whose chips and rows are live read as a
  // contradiction. Keeping this live is deliberate.
  test('a single degraded row shows immediately, undebounced', () => {
    const [group] = groupComponentResources(
      [
        resource({
          install_component_id: 'a',
          name: 'a-blip',
          health: 'degraded',
        }),
        resource({
          install_component_id: 'a',
          name: 'a-ok',
          health: 'healthy',
        }),
      ],
      { a: 'alpha' }
    )

    expect(group.worst).toBe('degraded')
    expect(group.failing).toBe(1)
    expect(group.live).toBe(2)
  })
})
