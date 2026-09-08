import { InstallOverviewCards } from './InstallOverviewCards'

export default {
  title: 'lite/organisms/InstallOverviewCards',
}

export const Default = () => (
  <div className="p-8">
    <InstallOverviewCards
      install={{
        id: 'inst_prod',
        name: 'production',
        composite_health_status: 'healthy',
        composite_health_status_description: 'All components are healthy.',
        drifted_objects: [],
        app_branch: {
          id: 'br_main',
          name: 'main',
          configs: [{ config_number: 14 }],
        },
      }}
      latestSync={{
        id: 'sync_latest',
        created_at: '2026-09-08T13:45:00Z',
        app_branch_id: 'br_main',
        triggered_by: 'git',
        vcs_connection_commit: {
          sha: 'a1b2c3d4e5f6',
          message: 'Update component versions',
          author_name: 'Alex Smith',
        },
      }}
    />
  </div>
)

export const NeedsAttention = () => (
  <div className="p-8">
    <InstallOverviewCards
      install={{
        id: 'inst_prod',
        name: 'production',
        composite_health_status: 'degraded',
        composite_health_status_description: 'One component is unhealthy.',
        drifted_objects: [
          { target_id: 'cmp_api' },
          { target_id: 'cmp_worker' },
        ],
        app_branch: { id: 'br_main', name: 'main' },
      }}
    />
  </div>
)

export const Loading = () => (
  <div className="p-8">
    <InstallOverviewCards isLoading />
  </div>
)
