export default {
  title: 'Features / Branches / Template changes',
}

import type { DiffSectionData } from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'
import { ConfigChangesViewer } from './ConfigChangesViewer'
import { TemplateChangesButton } from './TemplateChangesButton'

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

export const Button = () => (
  <TemplateChangesButton
    summary={{ added: 1, removed: 1, changed: 1 }}
    sections={sections}
    versionLabel="v13 → v14"
    previousSha="9f8e7d6c5b4a"
    sha="a1b2c3d4e5f6"
  />
)

export const NoChanges = () => <ConfigChangesViewer sections={[]} />
NoChanges.storyName = 'No changes'

export const Pending = () => <TemplateChangesButton sections={[]} isPending />
