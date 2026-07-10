import { BranchConnectionStep } from './BranchConnectionStep'

export default {
  title: 'Installs/BranchConnectionStep',
}

const mockBranches: any[] = [
  {
    id: 'branch-1',
    name: 'main',
    configs: [
      {
        install_groups: [
          {
            id: 'group-1',
            name: 'Staging',
            label_selector: { match_labels: { env: 'staging' } },
            install_ids: [],
          },
          {
            id: 'group-2',
            name: 'Production',
            label_selector: { match_labels: { env: 'production' } },
            install_ids: [],
          },
        ],
      },
    ],
  },
  {
    id: 'branch-2',
    name: 'release-v2',
    configs: [
      {
        install_groups: [
          {
            id: 'group-3',
            name: 'Canary',
            install_ids: ['inst-001', 'inst-002'],
          },
        ],
      },
    ],
  },
]

export const WithBranches = () => (
  <div className="max-w-xl p-4">
    <BranchConnectionStep
      branches={mockBranches}
      installId="inst-new"
      orgId="org123"
      onDone={() => alert('done')}
    />
  </div>
)

export const NoBranches = () => (
  <div className="max-w-xl p-4">
    <BranchConnectionStep
      branches={[]}
      installId="inst-new"
      orgId="org123"
      onDone={() => alert('done')}
    />
  </div>
)
