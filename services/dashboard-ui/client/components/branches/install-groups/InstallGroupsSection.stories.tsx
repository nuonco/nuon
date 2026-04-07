import { InstallGroupsSection } from './InstallGroupsSection'

export default { title: 'Branches/InstallGroups/InstallGroupsSection' }

const installsById = {
  'inst-1': { id: 'inst-1', name: 'Production US East', status_v2: { status: 'installed' } },
  'inst-2': { id: 'inst-2', name: 'Production EU West', status_v2: { status: 'installed' } },
  'inst-3': { id: 'inst-3', name: 'Staging', status_v2: { status: 'deploying' } },
} as any

export const Default = () => (
  <InstallGroupsSection
    config={{
      install_groups: [
        {
          id: 'group-1',
          name: 'Canary',
          max_parallel: 1,
          requires_approval: true,
          rollback_on_failure: false,
          install_ids: ['inst-1'],
        },
        {
          id: 'group-2',
          name: 'Production',
          max_parallel: 3,
          requires_approval: false,
          rollback_on_failure: true,
          install_ids: ['inst-2', 'inst-3'],
        },
      ],
    } as any}
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
