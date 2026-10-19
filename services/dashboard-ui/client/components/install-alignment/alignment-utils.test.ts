import { describe, expect, test } from 'bun:test'
import {
  alignmentResource,
  openDeploymentResource,
  stateDriftResource,
  unappliedResource,
} from '@/components/__fixtures__/install-alignment'
import {
  branchRunHref,
  driftCounts,
  filterDrifted,
  isDrifted,
} from './alignment-utils'

const open = openDeploymentResource({ id: 'ic-open' })
const none = unappliedResource({ id: 'ic-none' })
const scan = stateDriftResource({ id: 'ic-scan' })
const drifted = [open, none, scan]

describe('alignment utils', () => {
  test('only drift categories count as drifted', () => {
    expect(isDrifted(none)).toBe(true)
    expect(isDrifted(scan)).toBe(true)
    expect(isDrifted(alignmentResource({ category: 'unknown' }))).toBe(false)
    expect(isDrifted(alignmentResource({ category: 'disabled' }))).toBe(false)
    expect(isDrifted(alignmentResource({ category: 'torn-down' }))).toBe(false)
  })

  test('category filters keep their matching resources', () => {
    expect(filterDrifted(drifted, 'all')).toEqual(drifted)
    expect(filterDrifted(drifted, 'changes-not-applied')).toEqual([open, none])
    expect(filterDrifted(drifted, 'state-drift-detected')).toEqual([scan])
  })

  test('counts split unapplied changes and state drift', () => {
    expect(driftCounts(drifted)).toEqual({
      all: 3,
      unapplied: 2,
      stateDrift: 1,
    })
  })

  test('branch run links use the workflow ID', () => {
    expect(
      branchRunHref({
        orgId: 'org-1',
        appId: 'app-1',
        commit: { branch_id: 'br-1', workflow_id: 'wf-1', run_id: 'abr-1' },
      })
    ).toBe('/org-1/apps/app-1/branches/br-1/runs/wf-1')
    expect(
      branchRunHref({ orgId: 'org-1', appId: 'app-1', commit: { run_id: 'x' } })
    ).toBeUndefined()
  })
})
