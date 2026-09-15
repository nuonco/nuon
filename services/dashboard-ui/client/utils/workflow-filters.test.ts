import { describe, expect, test } from 'bun:test'
import { DateTime } from 'luxon'
import {
  datePresetFilterParameter,
  previewFilterParameter,
  readWorkflowFilters,
  statusFilterParameter,
  typeFilterParameter,
  workflowTypeOptions,
} from './workflow-filters'

describe('workflow status filters', () => {
  test('expands groups and removes duplicate statuses', () => {
    const statuses = statusFilterParameter(['running', 'failed'])?.split(',')

    expect(statuses).toContain('in-progress')
    expect(statuses).toContain('error')
    expect(
      statuses?.filter((status) => status === 'failed-pending-retry')
    ).toHaveLength(1)
  })
})

describe('workflow type filters', () => {
  test('expands groups to API workflow types', () => {
    expect(typeFilterParameter(['deploy'])).toBe(
      'manual_deploy,deploy_components'
    )
    expect(typeFilterParameter(['branch-manual'])).toBe(
      'app_branches_manual_update'
    )
  })

  test('separates install and app options', () => {
    expect(workflowTypeOptions('install')).toContain('drift')
    expect(workflowTypeOptions('install')).not.toContain('branch-manual')
    expect(workflowTypeOptions('app')).toContain('branch-manual')
    expect(workflowTypeOptions('app')).not.toContain('drift')
  })
})

describe('workflow preview filters', () => {
  test('only constrains the API when exactly one option is selected', () => {
    expect(previewFilterParameter(['preview'])).toBe(true)
    expect(previewFilterParameter(['rollout'])).toBe(false)
    expect(previewFilterParameter(['preview', 'rollout'])).toBeUndefined()
    expect(previewFilterParameter([])).toBeUndefined()
  })
})

describe('workflow date filters', () => {
  test('uses the widest selected date range', () => {
    const now = DateTime.fromISO('2026-09-15T12:00:00Z')

    expect(datePresetFilterParameter(['24h', '7d'], now)).toBe(
      '2026-09-08T12:00:00.000Z'
    )
  })
})

describe('workflow URL filters', () => {
  test('reads grouped app filters and converts them to API parameters', () => {
    const filters = readWorkflowFilters(
      new URLSearchParams(
        'q=deploy&status=running&type=branch-manual&preview=preview'
      ),
      'app'
    )

    expect(filters.api.search).toBe('deploy')
    expect(filters.api.status).toContain('in-progress')
    expect(filters.api.type).toBe('app_branches_manual_update')
    expect(filters.api.preview).toBe(true)
    expect(filters.filtered).toBe(true)
  })

  test('accepts the previous install search parameter', () => {
    const filters = readWorkflowFilters(
      new URLSearchParams('search=provision'),
      'install'
    )

    expect(filters.search).toBe('provision')
  })
})
