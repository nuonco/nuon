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

  test('uses whole-file entity content when there are no field diffs', () => {
    const changes = configChanges([
      section({
        name: 'Actions',
        sectionKey: 'actions',
        grouped: true,
        entities: [
          {
            name: 'healthcheck',
            op: 'change',
            fields: [],
            content: {
              op: 'change',
              before: "timeout = '1m'\n",
              after: "timeout = '2m'\n",
            },
          },
        ],
      }),
    ])

    expect(changes.sections).toHaveLength(1)
    expect(changes.sections[0]).toMatchObject({
      title: 'healthcheck',
      before: "timeout = '1m'\n",
      after: "timeout = '2m'\n",
    })
    expect(changes.sections[0].error).toBeUndefined()
  })

  test('attaches builds to the matching component and sandbox rows', () => {
    const changes = configChanges(
      [
        section({
          name: 'Components',
          sectionKey: 'components',
          grouped: true,
          entities: [
            {
              name: 'acme-image',
              op: 'change',
              changeKinds: ['source'],
              fields: [
                { key: 'source', op: 'change', diff: 'source files changed' },
              ],
            },
            {
              name: 'acme-api',
              op: 'change',
              fields: [
                {
                  key: 'namespace',
                  op: 'change',
                  diff: "'legacy' -> 'acme'",
                },
              ],
            },
            {
              name: 'acme-worker',
              op: 'change',
              changeKinds: ['source'],
              fields: [],
            },
          ],
        }),
      ],
      [],
      [
        {
          id: 'bld-1',
          name: 'acme-image',
          kind: 'component',
          status: 'active',
          changeReason: 'source_changed',
        },
        {
          id: 'bld-2',
          name: 'acme-api',
          kind: 'component',
          status: 'active',
          changeReason: 'source_and_config',
        },
        {
          id: 'sbld-1',
          name: 'Sandbox',
          kind: 'sandbox',
          status: 'active',
          changeReason: 'source_changed',
          commit: { sha: 'abc1234', message: 'Widen the acme VPC' },
        },
      ]
    )

    expect(changes.sections.map(({ id, group }) => [id, group])).toEqual([
      ['components/acme-image', 'Components'],
      ['components/acme-api', 'Components'],
      ['sandbox', 'Sandbox'],
    ])
    expect(changes.sections[0]).toMatchObject({
      before: '',
      after: '',
      build: { id: 'bld-1' },
    })
    expect(changes.sections[0].error).toBeUndefined()
    expect(changes.sections[1].build?.id).toBe('bld-2')
    expect(changes.sections[1].after).toContain('namespace = "acme"')
    expect(changes.sections[2].searchable).toContain('Widen the acme VPC')
    expect(changes.summary.update).toBe(3)
  })

  test('adds a row for a rebuild that has no template diff', () => {
    const changes = configChanges(
      [],
      [],
      [
        {
          id: 'bld-1',
          name: 'acme-image',
          kind: 'component',
          status: 'active',
          changeReason: 'source_changed',
        },
      ]
    )

    expect(changes.sections).toHaveLength(1)
    expect(changes.sections[0]).toMatchObject({
      id: 'components/acme-image',
      group: 'Components',
      build: { id: 'bld-1' },
    })
  })
})
