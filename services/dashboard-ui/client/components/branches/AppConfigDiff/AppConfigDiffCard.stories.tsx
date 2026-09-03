import type { DiffSectionData } from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'
import { AppConfigDiffCard } from './AppConfigDiffCard'

export default {
  title: 'Branches/AppConfigDiffCard',
}

const componentsSection: DiffSectionData = {
  name: 'Components',
  sectionKey: 'components',
  additions: 1,
  removals: 0,
  changed: 1,
  grouped: true,
  entities: [
    {
      name: 'api',
      op: 'add',
      componentType: 'docker_build',
      fields: [{ key: 'image', op: 'add', diff: '+ image: api:latest' }],
    },
    {
      name: 'ec2',
      op: 'change',
      componentType: 'terraform_module',
      fields: [
        {
          key: 'instance_type',
          op: 'change',
          diff: '- instance_type: t3.small\n+ instance_type: t3.medium',
        },
      ],
    },
  ],
  fields: [],
}

export const WithChanges = () => (
  <div className="max-w-3xl">
    <AppConfigDiffCard
      sections={[componentsSection]}
      summary={{ added: 1, removed: 0, changed: 1 }}
      versionLabel="v5 → v6"
    />
  </div>
)

export const NoChanges = () => (
  <div className="max-w-3xl">
    <AppConfigDiffCard
      sections={[]}
      summary={{ added: 0, removed: 0, changed: 0 }}
      versionLabel="v6"
    />
  </div>
)

export const Loading = () => (
  <div className="max-w-3xl">
    <AppConfigDiffCard
      sections={[]}
      summary={null}
      isLoading
      versionLabel="v6"
    />
  </div>
)

export const Snapshot = () => (
  <div className="max-w-3xl">
    <AppConfigDiffCard
      title="Config"
      presentation="snapshot"
      sections={[
        {
          ...componentsSection,
          additions: 2,
          removals: 0,
          changed: 0,
          entities: componentsSection.entities.map((e) => ({
            ...e,
            op: 'add' as const,
            fields: e.fields.map((f) => ({ ...f, op: 'add' })),
          })),
        },
      ]}
      summary={null}
    />
  </div>
)

export const Collapsed = () => (
  <div className="max-w-3xl">
    <AppConfigDiffCard
      sections={[componentsSection]}
      summary={{ added: 1, removed: 0, changed: 1 }}
      versionLabel="v5 → v6"
      isOpen={false}
    />
  </div>
)
