import { describe, expect, test } from 'bun:test'
import {
  filterInstalls,
  installStatusBucket,
  statusSlices,
  withQueuedInstalls,
} from './install-status'

describe('install status', () => {
  test('buckets workflow statuses into the rollout slices', () => {
    expect(installStatusBucket('succeeded')).toBe('success')
    expect(installStatusBucket('failed')).toBe('error')
    expect(installStatusBucket('approval-awaiting')).toBe('approval-awaiting')
    expect(installStatusBucket('user-skipped')).toBe('cancelled')
    expect(installStatusBucket(undefined)).toBe('pending')
  })

  test('counts slices in display order and drops empty ones', () => {
    expect(
      statusSlices([
        { status: 'pending' },
        { status: 'success' },
        { status: 'success' },
        { status: 'in-progress' },
      ]).map((slice) => [slice.status, slice.count])
    ).toEqual([
      ['success', 2],
      ['in-progress', 1],
      ['pending', 1],
    ])
  })

  test('pads a group up to its planned install count', () => {
    expect(
      withQueuedInstalls(
        [{ id: 'ins-1', name: 'alpha', status: 'success' }],
        3
      ).map((install) => install.status)
    ).toEqual(['success', 'pending', 'pending'])
  })

  test('filters installs by name and status', () => {
    const installs = [
      { id: '1', name: 'alpha', status: 'success' },
      { id: '2', name: 'alpine', status: 'error' },
      { id: '3', name: 'bravo', status: 'success' },
    ]
    expect(
      filterInstalls(installs, { query: 'alp', status: 'success' }).map(
        (install) => install.id
      )
    ).toEqual(['1'])
  })
})
