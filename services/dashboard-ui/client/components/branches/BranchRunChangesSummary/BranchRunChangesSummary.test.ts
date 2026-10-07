import { describe, expect, test } from 'bun:test'
import type { DiffSectionData } from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'
import {
  summarySectionsFromComparisonConfigDiff,
  withBuildChangeKinds,
} from './BranchRunChangesSummary'

const components = (
  entries: DiffSectionData['entities']
): DiffSectionData => ({
  name: 'Components',
  sectionKey: 'components',
  additions: 1,
  removals: 0,
  changed: 0,
  grouped: true,
  entities: entries,
  fields: [],
})

describe('withBuildChangeKinds', () => {
  test('marks config rows and adds a source-only component', () => {
    const sections = summarySectionsFromComparisonConfigDiff({
      additions: 1,
      removals: 1,
      changed: 0,
      sections: [
        {
          name: 'Components',
          additions: 1,
          removals: 0,
          changed: 0,
          entries: [
            { op: 'add', name: 'img_nginx', description: 'container image' },
          ],
        },
        {
          name: 'Actions',
          additions: 0,
          removals: 1,
          changed: 0,
          entries: [{ op: 'remove', name: 'migrate' }],
        },
      ],
    })

    const next = withBuildChangeKinds(sections, [
      {
        component_id: 'cmp_nginx',
        component_name: 'img_nginx',
        change_reason: 'config_changed',
      },
      {
        component_id: 'cmp_alb',
        component_name: 'alb',
        change_reason: 'source_changed',
      },
    ])

    const componentSection = next.find(
      (section) => section.sectionKey === 'components'
    )
    expect(
      componentSection?.entities.map((entity) => [entity.name, entity.changeKinds])
    ).toEqual([
      ['img_nginx', ['config']],
      ['alb', ['source']],
    ])
    expect(componentSection?.changed).toBe(1)
  })

  test('marks a component that changed source and config with both', () => {
    const next = withBuildChangeKinds(
      [
        components([
          {
            name: 'img_nginx',
            op: 'add',
            fields: [{ key: 'change', op: 'add', diff: 'image' }],
          },
        ]),
      ],
      [
        {
          component_name: 'img_nginx',
          change_reason: 'source_and_config',
        },
      ]
    )

    expect(next[0]?.entities[0]?.changeKinds).toEqual(['source', 'config'])
  })
})
