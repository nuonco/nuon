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
  listedResources,
  resourcePage,
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

  test('listed resources page newest first, only within the selected config, filtering before slicing', () => {
    const listed = (index: number, overrides = {}) =>
      alignmentResource({
        id: `ic-${index}`,
        name: `service-${index}`,
        component_id: `cmp-${index}`,
        expected_config_id: 'cfg-selected',
        ...overrides,
      })
    const resources = [
      ...Array.from({ length: 12 }, (_, index) => listed(index)),
      listed(12, { category: 'disabled' }),
      listed(13, { component_type: 'terraform_module' }),
      listed(14, { expected_config_id: undefined, category: 'unknown' }),
      listed(15, { category: 'changes-not-applied' }),
      listed(16, {
        open_deployment: { id: 'wf-open', status: 'in-progress' },
      }),
      listed(17, { kind: 'image', component_type: 'external_image' }),
    ]
    const ids = (items: typeof resources) => items.map((item) => item.id)

    const all = listedResources(resources, ['component'])
    expect(ids(all)).toEqual(
      [13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0].map((i) => `ic-${i}`)
    )
    expect(resourcePage(all, 0, 10)).toEqual({
      offset: 0,
      ids: ids(all.slice(0, 10)),
      hasNext: true,
    })
    expect(resourcePage(all, 10, 10)).toEqual({
      offset: 10,
      ids: ['ic-3', 'ic-2', 'ic-1', 'ic-0'],
      hasNext: false,
    })
    expect(resourcePage(all, 20, 10).offset).toBe(0)
    expect(resourcePage([], 10, 10)).toEqual({
      offset: 0,
      ids: [],
      hasNext: false,
    })

    expect(
      ids(listedResources(resources, ['component'], { enabled: 'disabled' }))
    ).toEqual(['ic-12'])
    expect(
      listedResources(resources, ['component'], { enabled: 'enabled' })
    ).toHaveLength(13)
    expect(
      ids(
        listedResources(resources, ['component'], {
          types: ['terraform_module'],
        })
      )
    ).toEqual(['ic-13'])
    expect(
      ids(listedResources(resources, ['component'], { search: 'SERVICE-1' }))
    ).toEqual(['ic-13', 'ic-12', 'ic-11', 'ic-10', 'ic-1'])
    expect(
      ids(listedResources(resources, ['component'], { search: 'cmp-3' }))
    ).toEqual(['ic-3'])
    expect(
      listedResources(resources, ['component'], { search: 'cmp-' })
    ).toHaveLength(0)
    expect(ids(listedResources(resources, ['image']))).toEqual(['ic-17'])
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
