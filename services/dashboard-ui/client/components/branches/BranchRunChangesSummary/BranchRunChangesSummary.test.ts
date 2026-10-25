import { describe, expect, test } from 'bun:test'
import type { DiffSectionData } from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'
import type { TCompositeError } from '@/types'
import {
  overlaySectionDetail,
  summarySectionsFromComparisonConfigDiff,
  withBuildChangeKinds,
} from './BranchRunChangesSummary'
import { configDiagnosticLines } from './config-diagnostics'

const components = (entries: DiffSectionData['entities']): DiffSectionData => ({
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
      componentSection?.entities.map((entity) => [
        entity.name,
        entity.changeKinds,
      ])
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

describe('summarySectionsFromComparisonConfigDiff', () => {
  test('surfaces connected_repo source files as file entries', () => {
    const sections = summarySectionsFromComparisonConfigDiff({
      additions: 0,
      removals: 0,
      changed: 1,
      sections: [
        {
          name: 'Components',
          additions: 0,
          removals: 0,
          changed: 1,
          entries: [
            {
              op: 'change',
              name: 'grafana_lgtm',
              source_changed: true,
              source_files: [
                {
                  path: 'byoc-nuon/components/grafana_lgtm/values.yaml',
                  diff: '+ securityPolicy: ""',
                },
              ],
            },
          ],
        },
      ],
    })

    const grafana = sections[0]?.entities[0]
    expect(grafana?.fields).toEqual([
      { key: 'source', op: 'change', diff: 'source files changed' },
    ])
    expect(grafana?.files).toEqual([
      {
        name: 'byoc-nuon/components/grafana_lgtm/values.yaml',
        op: 'change',
        patch: '+ securityPolicy: ""',
      },
    ])
  })
})

describe('overlaySectionDetail', () => {
  test('keeps connected_repo source files alongside the detailed diff files', () => {
    const summary = summarySectionsFromComparisonConfigDiff({
      additions: 0,
      removals: 0,
      changed: 1,
      sections: [
        {
          name: 'Components',
          additions: 0,
          removals: 0,
          changed: 1,
          entries: [
            {
              op: 'change',
              name: 'grafana_lgtm',
              source_changed: true,
              source_files: [
                {
                  path: 'byoc-nuon/components/grafana_lgtm/values.yaml',
                  diff: '+ securityPolicy: ""',
                },
              ],
            },
          ],
        },
      ],
    })

    const detailed: DiffSectionData[] = [
      {
        name: 'Components',
        sectionKey: 'components',
        additions: 0,
        removals: 0,
        changed: 1,
        grouped: true,
        fields: [],
        entities: [
          {
            name: 'grafana_lgtm',
            op: 'change',
            fields: [{ key: 'branch', op: 'change', diff: "main -> jm/x" }],
            files: [{ name: 'values_files', op: 'change', before: 'a', after: 'b' }],
          },
        ],
      },
    ]

    const merged = overlaySectionDetail(summary, detailed)
    const grafana = merged[0]?.entities[0]
    expect(grafana?.fields).toEqual([{ key: 'branch', op: 'change', diff: 'main -> jm/x' }])
    expect(grafana?.files).toEqual([
      { name: 'values_files', op: 'change', before: 'a', after: 'b' },
      {
        name: 'byoc-nuon/components/grafana_lgtm/values.yaml',
        op: 'change',
        patch: '+ securityPolicy: ""',
      },
    ])
  })
})

describe('configDiagnosticLines', () => {
  test('splits the validation code section into lines', () => {
    const error = {
      type: 'app_branch_run.config_validation_failed',
      message: 'App configuration validation failed',
      sections: [
        { kind: 'markdown', heading: 'Why', body: 'The branch run stopped.' },
        {
          kind: 'code',
          heading: 'Validation errors',
          body: 'components/api.toml: toml: line 4: expected key\n\ncomponents/worker.toml: image is required\n',
        },
      ],
    } as TCompositeError

    expect(configDiagnosticLines(error)).toEqual([
      'components/api.toml: toml: line 4: expected key',
      'components/worker.toml: image is required',
    ])
  })

  test('ignores other composite errors', () => {
    expect(
      configDiagnosticLines({
        type: 'install_group.install_update_failed',
        sections: [{ kind: 'code', body: 'deploy failed' }],
      } as TCompositeError)
    ).toEqual([])
  })
})
