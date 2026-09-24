import { describe, expect, test } from 'bun:test'
import { DateTime } from 'luxon'
import type { TWorkflowType } from '@/types/ctl-api.types'
import {
  WORKFLOW_STATUS_GROUPS,
  WORKFLOW_TYPE_GROUPS,
  createdAtGteForPreset,
  datePresetFilterParameter,
  previewFilterParameter,
  statusFilterParameter,
  typeFilterParameter,
  workflowTypeOptions,
  workflowTypesForGroups,
} from './workflow-filters'

const ALL_TYPES = Object.keys(WORKFLOW_TYPE_GROUPS) as TWorkflowType[]

describe('status filter', () => {
  test('one option expands to its statuses', () => {
    expect(statusFilterParameter(['failed'])).toBe('error,failed-pending-retry')
    expect(statusFilterParameter(['queued'])).toBe('pending,queued')
  })

  test('several options union', () => {
    expect(statusFilterParameter(['succeeded', 'cancelled'])).toBe(
      'success,cancelled'
    )
  })

  test('a status in two groups is sent once', () => {
    const statuses = statusFilterParameter(['failed', 'running'])?.split(',')
    expect(statuses).toContain('failed-pending-retry')
    expect(
      statuses?.filter((status) => status === 'failed-pending-retry')
    ).toHaveLength(1)
  })

  test('running covers every in-flight status', () => {
    expect(statusFilterParameter(['running'])?.split(',')).toEqual([
      'in-progress',
      'planning',
      'applying',
      'checking-plan',
      'retrying',
      'failed-pending-retry',
    ])
  })

  test('nothing selected sends no parameter', () => {
    expect(statusFilterParameter([])).toBeUndefined()
  })

  test('every option maps to at least one status', () => {
    for (const statuses of Object.values(WORKFLOW_STATUS_GROUPS)) {
      expect(statuses.length).toBeGreaterThan(0)
    }
  })
})

describe('type filter', () => {
  test('one option expands to its types', () => {
    expect(typeFilterParameter(['drift'])).toBe(
      'drift_run,drift_run_reprovision_sandbox'
    )
    expect(typeFilterParameter(['branch-manual'])).toBe(
      'app_branches_manual_update'
    )
  })

  test('several options union', () => {
    expect(
      typeFilterParameter(['branch-config-repo', 'branch-component-repo'])
    ).toBe(
      'app_branches_config_repo_update,app_branches_component_repo_update'
    )
  })

  test('nothing selected sends no parameter', () => {
    expect(typeFilterParameter([])).toBeUndefined()
  })

  test('every workflow type has a group', () => {
    for (const type of ALL_TYPES) {
      expect(WORKFLOW_TYPE_GROUPS[type]).toBeTruthy()
    }
  })

  test('app options exclude install-owned types and vice versa', () => {
    const appTypes = workflowTypesForGroups(workflowTypeOptions('app'))
    const installTypes = workflowTypesForGroups(workflowTypeOptions('install'))

    expect(appTypes).toContain('app_branches_manual_update')
    expect(appTypes).not.toContain('drift_run')
    expect(installTypes).toContain('drift_run')
    expect(installTypes).not.toContain('app_branches_manual_update')
    expect(
      appTypes.filter((type) => installTypes.includes(type))
    ).toHaveLength(0)
  })

  test('the two owners together cover every type', () => {
    const covered = new Set([
      ...workflowTypesForGroups(workflowTypeOptions('app')),
      ...workflowTypesForGroups(workflowTypeOptions('install')),
    ])
    expect([...covered].sort()).toEqual([...ALL_TYPES].sort())
  })
})

describe('preview filter', () => {
  test('preview alone asks for preview runs only', () => {
    expect(previewFilterParameter(['preview'])).toBe(true)
  })

  test('rollout alone asks for rollout runs only', () => {
    expect(previewFilterParameter(['rollout'])).toBe(false)
  })

  test('neither selected sends no parameter', () => {
    expect(previewFilterParameter([])).toBeUndefined()
  })

  test('both selected sends no parameter', () => {
    expect(previewFilterParameter(['preview', 'rollout'])).toBeUndefined()
  })
})

describe('date presets', () => {
  const now = DateTime.fromISO('2026-09-10T12:00:00Z', { zone: 'utc' })

  test('each preset resolves to its lower bound', () => {
    expect(createdAtGteForPreset('24h', now)).toBe('2026-09-09T12:00:00.000Z')
    expect(createdAtGteForPreset('7d', now)).toBe('2026-09-03T12:00:00.000Z')
    expect(createdAtGteForPreset('30d', now)).toBe('2026-08-11T12:00:00.000Z')
  })

  test('no preset sends no bound', () => {
    expect(createdAtGteForPreset(undefined, now)).toBeUndefined()
    expect(datePresetFilterParameter([], now)).toBeUndefined()
  })

  test('several presets resolve to the widest window', () => {
    expect(datePresetFilterParameter(['24h', '30d'], now)).toBe(
      '2026-08-11T12:00:00.000Z'
    )
  })
})
