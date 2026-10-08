import { describe, expect, test } from 'bun:test'
import type { DiffSectionData } from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'
import { configChanges } from './config-changes'

const section = (
  overrides: Partial<DiffSectionData> &
    Pick<DiffSectionData, 'name' | 'sectionKey'>
): DiffSectionData => ({
  additions: 0,
  removals: 0,
  changed: 0,
  grouped: false,
  entities: [],
  fields: [],
  ...overrides,
})

describe('configChanges', () => {
  test('turns grouped entities into searchable config sections', () => {
    const changes = configChanges([
      section({
        name: 'Components',
        sectionKey: 'components',
        grouped: true,
        additions: 1,
        entities: [
          {
            name: 'cache',
            op: 'add',
            componentType: 'helm_chart',
            fields: [
              { key: 'chart_name', op: 'add', diff: "'cache'" },
              {
                key: 'namespace',
                op: 'change',
                diff: "'legacy' -> 'acme'",
              },
            ],
          },
        ],
      }),
    ])

    expect(changes.sections).toHaveLength(1)
    expect(changes.sections[0]).toMatchObject({
      title: 'cache',
      group: 'Components',
      operation: 'create',
      kind: 'config',
    })
    expect(changes.sections[0].before).toContain('namespace = "legacy"')
    expect(changes.sections[0].after).toContain('chart_name = "cache"')
    expect(changes.summary.create).toBe(1)
  })

  test('adds source files as their own group and ignores unchanged files', () => {
    const changes = configChanges(
      [],
      [
        {
          path: 'values/cache.yaml',
          kind: 'helm values',
          change: 'added',
          after: 'replicaCount: 2\n',
        },
        {
          path: 'values/api.yaml',
          kind: 'helm values',
          change: 'unchanged',
          before: 'same\n',
          after: 'same\n',
        },
      ]
    )

    expect(changes.sections.map((item) => item.title)).toEqual([
      'values/cache.yaml',
    ])
    expect(changes.sections[0].group).toBe('Source files')
    expect(changes.summary.create).toBe(0)
  })
})
