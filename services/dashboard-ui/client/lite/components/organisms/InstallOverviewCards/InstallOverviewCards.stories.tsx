import { ComponentDocs } from '../../__stories__/ComponentDocs'
import { InstallOverviewCards } from './InstallOverviewCards'

export default {
  title: 'lite/organisms/InstallOverviewCards',
}

export const Overview = () => (
  <ComponentDocs
    name="InstallOverviewCards"
    tier="organism"
    summary="The install overview for current health, expected app-branch update, and running services."
    use={[
      'Place at the top of the install overview, above the rest of the page.',
      'Use the loading state while the install and its latest branch update resolve.',
    ]}
    avoid={[
      'Do not add app-branch history here; use the Activity timeline.',
      'Do not promote every install status into an equal card.',
    ]}
    rules={[
      'Install status combines lifecycle, health, and a secondary drift summary.',
      'Expected update identifies the app-branch run and commit intended for this install.',
      'Runner, sandbox, and components remain separate status facets.',
    ]}
    props={[
      {
        name: 'install',
        type: 'TInstall',
        description: 'Install whose lifecycle, health, and services are summarized.',
      },
      {
        name: 'lastBranchUpdate',
        type: 'IInstallBranchUpdate',
        description: 'Most recent app-branch run expected on this install.',
      },
      {
        name: 'loading',
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
        lifecycle_phase: { phase: 'active' },
        runner_status: 'active',
        sandbox_status: 'active',
        composite_component_status: 'active',
        drifted_objects: [],
        app_branch: {
          id: 'br_main',
          name: 'main',
          configs: [{ config_number: 14 }],
        },
      }}
      lastBranchUpdate={{
        runId: 'run_01k4m8f6a9',
        runHref: '/org_example/apps/app_payments/branches/br_main/activity',
        branchName: 'main',
        status: 'success',
        updatedAt: '2026-09-08T13:45:00Z',
        commit: {
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
        lifecycle_phase: { phase: 'active' },
        runner_status: 'active',
        sandbox_status: 'degraded',
        sandbox_health_status: 'degraded',
        composite_component_status: 'error',
        drifted_objects: [
          { target_id: 'cmp_api' },
          { target_id: 'cmp_worker' },
        ],
        app_branch: { id: 'br_main', name: 'main' },
      }}
      lastBranchUpdate={{
        runId: 'run_01k4m8f6a9',
        branchName: 'main',
        status: 'error',
        commit: {
          sha: 'b7c8d9e0f1a2',
          message: 'Roll out worker update',
        },
      }}
    />
  </div>
)

export const Provisioning = () => (
  <div className="p-8">
    <InstallOverviewCards
      install={{
        id: 'inst_prod',
        name: 'production',
        lifecycle_phase: {
          phase: 'provisioning',
          description: 'Provisioning the install sandbox.',
        },
        runner_status: 'active',
        sandbox_status: 'provisioning',
        composite_component_status: 'pending',
        app_branch: { id: 'br_main', name: 'main' },
      }}
      lastBranchUpdate={{
        runId: 'run_01k4m8f6a9',
        branchName: 'main',
        status: 'pending',
        commit: {
          sha: 'a1b2c3d4e5f6',
          message: 'Create production install',
        },
      }}
    />
  </div>
)

export const Deprovisioning = () => (
  <div className="p-8">
    <InstallOverviewCards
      install={{
        id: 'inst_prod',
        name: 'production',
        lifecycle_phase: {
          phase: 'deprovisioning',
          description: 'Tearing down install services.',
        },
        runner_status: 'active',
        sandbox_status: 'active',
        composite_component_status: 'active',
        app_branch: { id: 'br_main', name: 'main' },
      }}
      lastBranchUpdate={{
        runId: 'run_01k4m8f6a9',
        branchName: 'main',
        status: 'success',
        commit: {
          sha: 'a1b2c3d4e5f6',
          message: 'Update component versions',
        },
      }}
    />
  </div>
)

export const NeverUpdated = () => (
  <div className="p-8">
    <InstallOverviewCards
      install={{
        id: 'inst_prod',
        name: 'production',
        lifecycle_phase: { phase: 'pending' },
        runner_status: 'pending',
        sandbox_status: 'pending',
        composite_component_status: 'pending',
        app_branch: { id: 'br_main', name: 'main' },
      }}
    />
  </div>
)

export const Loading = () => (
  <div className="p-8">
    <InstallOverviewCards loading />
  </div>
)
