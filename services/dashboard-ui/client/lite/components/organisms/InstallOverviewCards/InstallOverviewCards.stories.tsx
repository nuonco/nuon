import { ComponentDocs } from '../../__stories__/ComponentDocs'
import { InstallOverviewCards } from './InstallOverviewCards'

export default {
  title: 'lite/organisms/InstallOverviewCards',
}

export const Overview = () => (
  <ComponentDocs
    name="InstallOverviewCards"
    tier="organism"
    summary="The four-card summary at the top of an install overview: health, drift, branch, and last update."
    use={[
      'Place at the top of the install overview, above the rest of the page.',
      'Use the loading state while the install and its latest config sync resolve.',
    ]}
    avoid={[
      'Do not collapse health and drift into one status; they are independent axes.',
      'Do not report an unscanned install as having no drift.',
    ]}
    rules={[
      'Health and drift each distinguish a missing evaluation from a healthy result.',
      'Drift counts the install drifted objects and pluralizes the resource count.',
      'Last update reads the commit from the latest config sync, not the branch latest run.',
    ]}
    props={[
      {
        name: 'install',
        type: 'TInstall',
        description: 'Install whose health, drift, and branch are summarized.',
      },
      {
        name: 'latestSync',
        type: 'TInstallConfigSync',
        description: 'Most recent config sync, used for the last update card.',
      },
      {
        name: 'isLoading',
        type: 'boolean',
        default: 'false',
        description: 'Shows card loading shapes while keeping every title.',
      },
    ]}
  />
)

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
