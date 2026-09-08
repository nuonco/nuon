import { AppBranchOverviewCards } from './AppBranchOverviewCards'

export default {
  title: 'lite/organisms/AppBranchOverviewCards',
}

export const Default = () => (
  <div className="p-8">
    <AppBranchOverviewCards
      branch={{
        id: 'br_main',
        name: 'main',
        configs: [
          {
            config_number: 14,
            connected_github_vcs_config: {
              repo: 'acme/payments',
              branch: 'main',
            },
          },
        ],
        latest_run: {
          updated_at: '2026-09-08T13:45:00Z',
          vcs_connection_commit: {
            sha: 'a1b2c3d4e5f6',
            message: 'Update component versions',
            author_name: 'Alex Smith',
            created_at: '2026-09-08T13:42:00Z',
          },
        },
      }}
      installCount={12}
    />
  </div>
)

export const Empty = () => (
  <div className="p-8">
    <AppBranchOverviewCards branch={{ id: 'br_main', name: 'main' }} installCount={0} />
  </div>
)

export const Loading = () => (
  <div className="p-8">
    <AppBranchOverviewCards isLoading />
  </div>
)
