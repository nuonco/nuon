import { describe, expect, it } from 'bun:test'
import type { DiffSectionData } from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'
import { mergeBranchSections } from './merge-branch-sections'

const makeSection = (
  sectionKey: string,
  overrides: Partial<DiffSectionData> = {}
): DiffSectionData => ({
  name: sectionKey,
  sectionKey,
  additions: 0,
  removals: 0,
  changed: 0,
  grouped: false,
  entities: [],
  fields: [],
  ...overrides,
})

const makeEntity = (
  name: string,
  op: 'add' | 'remove' | 'change' = 'change'
) => ({
  name,
  op,
  fields: [],
})

describe('mergeBranchSections', () => {
  it('returns installSections unchanged when branchSections is empty', () => {
    const install = [makeSection('inputs')]
    expect(mergeBranchSections(install, [])).toEqual(install)
  })

  it('adds a missing section from branch (inputs)', () => {
    const install = [makeSection('components', { grouped: true, entities: [] })]
    const branch = [
      makeSection('components', { grouped: true, entities: [] }),
      makeSection('inputs', {
        fields: [
          { key: 'region', op: 'change', diff: 'us-east-1 -> eu-west-1' },
        ],
      }),
    ]
    const result = mergeBranchSections(install, branch)
    expect(result.map((s) => s.sectionKey)).toContain('inputs')
  })

  it('does not duplicate an existing section', () => {
    const install = [makeSection('inputs')]
    const branch = [makeSection('inputs')]
    const result = mergeBranchSections(install, branch)
    expect(result.filter((s) => s.sectionKey === 'inputs')).toHaveLength(1)
  })

  it('adds source-only component entities not already in install sections', () => {
    const install = [
      makeSection('components', {
        grouped: true,
        entities: [makeEntity('acme-api')],
      }),
    ]
    const branch = [
      makeSection('components', {
        grouped: true,
        entities: [makeEntity('acme-api'), makeEntity('acme-worker')],
      }),
    ]
    const result = mergeBranchSections(install, branch)
    const components = result.find((s) => s.sectionKey === 'components')!
    const names = components.entities.map((e) => e.name)
    expect(names).toContain('acme-api')
    expect(names).toContain('acme-worker')
    expect(names).toHaveLength(2)
  })

  it('does not double-count a component entity already in the install diff', () => {
    const install = [
      makeSection('components', {
        grouped: true,
        entities: [makeEntity('acme-api'), makeEntity('acme-worker')],
        changed: 2,
      }),
    ]
    const branch = [
      makeSection('components', {
        grouped: true,
        entities: [makeEntity('acme-api'), makeEntity('acme-worker')],
      }),
    ]
    const result = mergeBranchSections(install, branch)
    const components = result.find((s) => s.sectionKey === 'components')!
    expect(components.entities).toHaveLength(2)
    expect(components.changed).toBe(2)
  })

  it('creates a components section from branch if none in install', () => {
    const install: DiffSectionData[] = []
    const branch = [
      makeSection('components', {
        grouped: true,
        entities: [makeEntity('acme-worker')],
      }),
    ]
    const result = mergeBranchSections(install, branch)
    expect(result[0].sectionKey).toBe('components')
    expect(result[0].entities[0].name).toBe('acme-worker')
  })

  it('updates changed/additions/removals counts when merging component entities', () => {
    const install = [
      makeSection('components', {
        grouped: true,
        entities: [makeEntity('acme-api', 'change')],
        changed: 1,
        additions: 0,
        removals: 0,
      }),
    ]
    const branch = [
      makeSection('components', {
        grouped: true,
        entities: [makeEntity('acme-worker', 'add')],
      }),
    ]
    const result = mergeBranchSections(install, branch)
    const components = result.find((s) => s.sectionKey === 'components')!
    expect(components.changed).toBe(1)
    expect(components.additions).toBe(1)
    expect(components.removals).toBe(0)
    expect(components.entities).toHaveLength(2)
  })
})
