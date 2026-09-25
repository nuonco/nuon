import { InstallGroupsSection } from './InstallGroupsSection'

export default { title: 'Branches/InstallGroups/InstallGroupsSection' }

const installsById = {
  'inst-1': {
    id: 'inst-1',
    name: 'Production US East',
    labels: { env: 'production' },
    status_v2: { status: 'installed' },
  },
  'inst-2': {
    id: 'inst-2',
    name: 'Production EU West',
    labels: { env: 'production' },
    status_v2: { status: 'installed' },
  },
  'inst-3': {
    id: 'inst-3',
    name: 'Staging',
    labels: { env: 'staging' },
    status_v2: { status: 'deploying' },
  },
} as any

export const Default = () => (
  <InstallGroupsSection
    config={
      {
        install_groups: [
          {
            id: 'group-1',
            name: 'Canary',
            max_parallel: 1,
            label_selector: { match_labels: { env: 'staging' } },
          },
          {
            id: 'group-2',
            name: 'Production',
            max_parallel: 2,
            label_selector: { match_labels: { env: 'production' } },
          },
        ],
      } as any
    }
    installsById={installsById}
    orgId="org-1"
  />
)

export const WithDefaultGroup = () => (
  <InstallGroupsSection
    config={
      {
        install_groups: [
          {
            id: 'group-1',
            name: 'Canary',
            max_parallel: 1,
            label_selector: { match_labels: { env: 'staging' } },
          },
          {
            id: 'group-2',
            name: 'All others',
            max_parallel: 3,
            default: true,
          },
        ],
      } as any
    }
    installsById={installsById}
    orgId="org-1"
  />
)

export const Empty = () => (
  <InstallGroupsSection
    config={{ install_groups: [] } as any}
    installsById={{}}
    orgId="org-1"
  />
)
