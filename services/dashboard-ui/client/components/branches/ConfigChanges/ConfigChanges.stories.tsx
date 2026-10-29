export default {
  title: 'Features / Branches / Changes',
}

import type { DiffSectionData } from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'
import { ConfigChangesLoading } from './ConfigChangesLoading'
import { ConfigChangesViewer } from './ConfigChangesViewer'
import { TemplateChangesButton } from './TemplateChangesButton'
import type { TTemplateBuildChange } from './config-changes'

const sections: DiffSectionData[] = [
  {
    name: 'Components',
    sectionKey: 'components',
    additions: 1,
    removals: 1,
    changed: 1,
    grouped: true,
    fields: [],
    entities: [
      {
        name: 'cache',
        op: 'add',
        componentType: 'helm_chart',
        fields: [
          { key: 'chart_name', op: 'add', diff: "'cache'" },
          { key: 'namespace', op: 'add', diff: "'acme'" },
        ],
      },
      {
        name: 'api',
        op: 'change',
        componentType: 'helm_chart',
        fields: [
          {
            key: 'public_repo.branch',
            op: 'change',
            diff: "'release-13' -> 'release-14'",
          },
        ],
      },
      {
        name: 'legacy-redis',
        op: 'remove',
        componentType: 'helm_chart',
        fields: [{ key: 'chart_name', op: 'remove', diff: "'redis'" }],
      },
    ],
  },
]

export const Viewer = () => (
  <ConfigChangesViewer
    sections={sections}
    versionLabel="v13 → v14"
    previousSha="9f8e7d6c5b4a"
    sha="a1b2c3d4e5f6"
    files={[
      {
        path: 'values/cache.yaml',
        kind: 'helm values',
        change: 'added',
        after: 'replicaCount: 2\n',
      },
    ]}
  />
)

const builds: TTemplateBuildChange[] = [
  {
    id: 'bld-api',
    name: 'api',
    kind: 'component',
    status: 'active',
    changeReason: 'source_and_config',
    href: '#',
  },
  {
    id: 'bld-1',
    name: 'acme-image',
    kind: 'component',
    status: 'active',
    changeReason: 'source_changed',
    href: '#',
  },
  {
    id: 'sbld-1',
    name: 'Sandbox',
    kind: 'sandbox',
    status: 'active',
    changeReason: 'source_changed',
    href: '#',
    commit: {
      sha: '0a1b2c3d4e5f60718293a4b5c6d7e8f901234567',
      message: 'Widen the acme VPC CIDR range\n\nLeaves room for a second AZ.',
      author: 'Example Developer',
    },
  },
]

export const WithBuilds = () => (
  <ConfigChangesViewer
    sections={sections}
    builds={builds}
    versionLabel="v13 → v14"
    previousSha="9f8e7d6c5b4a"
    sha="a1b2c3d4e5f6"
  />
)
WithBuilds.storyName = 'With builds'

const embeddedToml = (enabled: 'false' | 'true', branch: string) =>
  [
    'source = ""',
    `contents = ${JSON.stringify(
      [
        '---',
        'apiVersion: v1',
        'kind: ConfigMap',
        'metadata:',
        '  name: acme',
        'data:',
        `  enabled: "${enabled}"`,
        '  endpoint: http://collector.example.com',
      ].join('\n')
    )}`,
    'path = ""',
    '',
    '[helm_chart.public_repo]',
    "repo = 'acme/example'",
    `branch = '${branch}'`,
  ].join('\n')

export const EmbeddedToml = () => (
  <ConfigChangesViewer
    sections={[
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
            name: 'collector',
            op: 'change',
            componentType: 'helm_chart',
            fields: [],
            content: {
              op: 'change',
              before: embeddedToml('false', 'release-13'),
              after: embeddedToml('true', 'release-14'),
            },
          },
        ],
      },
    ]}
    versionLabel="v13 → v14"
    previousSha="9f8e7d6c5b4a"
    sha="a1b2c3d4e5f6"
  />
)
EmbeddedToml.storyName = 'Embedded toml'

export const BuildsOnly = () => (
  <ConfigChangesViewer sections={[]} builds={builds} versionLabel="v14" />
)
BuildsOnly.storyName = 'Builds only'

export const Button = () => (
  <TemplateChangesButton
    sections={sections}
    builds={builds}
    versionLabel="v13 → v14"
    previousSha="9f8e7d6c5b4a"
    sha="a1b2c3d4e5f6"
  />
)

const manySections: DiffSectionData[] = [
  {
    name: 'Components',
    sectionKey: 'components',
    additions: 0,
    removals: 0,
    changed: 24,
    grouped: true,
    fields: [],
    entities: Array.from({ length: 24 }, (_, index) => ({
      name: `service-${String(index + 1).padStart(2, '0')}`,
      op: 'change' as const,
      componentType: 'helm_chart',
      fields: [
        {
          key: 'chart_version',
          op: 'change' as const,
          diff: `'1.${index}.0' -> '1.${index + 1}.0'`,
        },
        {
          key: 'namespace',
          op: 'change' as const,
          diff: "'acme' -> 'acme-prod'",
        },
      ],
    })),
  },
]

export const ManyChanges = () => (
  <ConfigChangesViewer
    sections={manySections}
    versionLabel="v13 → v14"
    previousSha="9f8e7d6c5b4a"
    sha="a1b2c3d4e5f6"
  />
)
ManyChanges.storyName = 'Many changes'

export const ManyChangesPanel = () => (
  <TemplateChangesButton
    sections={manySections}
    versionLabel="v13 → v14"
    previousSha="9f8e7d6c5b4a"
    sha="a1b2c3d4e5f6"
  />
)
ManyChangesPanel.storyName = 'Many changes panel'

export const NoChanges = () => <ConfigChangesViewer sections={[]} />
NoChanges.storyName = 'No changes'

export const Pending = () => <TemplateChangesButton sections={[]} isPending />

export const Loading = () => <ConfigChangesLoading />
