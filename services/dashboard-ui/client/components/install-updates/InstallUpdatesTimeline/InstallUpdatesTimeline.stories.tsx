import { InstallUpdatesTimeline } from './InstallUpdatesTimeline'
import type { TInstallUpdate } from '@/types'

export default {
  title: 'InstallUpdates/InstallUpdatesTimeline',
}

const day = 86400000

const mockUpdates: TInstallUpdate[] = [
  {
    id: 'upd-3',
    type: 'app_config',
    created_at: new Date(Date.now() - day).toISOString(),
    workflow_id: 'wkf-3',
    app_config: {
      version: {
        id: 'iacv-3',
        new_app_config_id: 'appcfg-3',
        status: { status: 'success' },
        app_branch_run_id: 'abr-3',
        app_branch_run: {
          pr_number: 42,
          app_branch: { id: 'branch-1', name: 'feat/add-cache' },
          vcs_connection_commit: {
            sha: 'a1b2c3d4e5f6',
            message: 'Add a cache component to the deployment plan',
            author_name: 'Jane Doe',
          },
        },
      },
      diff: {
        added: [],
        removed: [],
        unchanged: [],
        stack_changed: true,
        stack_impacts: ['permissions'],
        stack_impact_reasons: [
          { from: 'role.maintenance', edge: 'stack_render' },
        ],
        changed: [
          {
            component_id: 'cmp-api',
            component_name: 'api',
            component_type: 'helm_chart',
            impact_reasons: [
              { from: 'role.maintenance', edge: 'operation_role' },
            ],
          },
        ],
      },
    },
  },
  {
    id: 'upd-2',
    type: 'stack',
    created_at: new Date(Date.now() - day * 3).toISOString(),
    stack: {
      version_id: 'stkv-2',
      status: { status: 'in-progress' },
      run_type: 'apply',
    },
  },
  {
    id: 'upd-1',
    type: 'inputs',
    created_at: new Date(Date.now() - day * 9).toISOString(),
    inputs: { input_config_id: 'inpcfg-1', keys: ['license', 'region'] },
  },
  {
    id: 'upd-0',
    type: 'install_config',
    created_at: new Date(Date.now() - day * 14).toISOString(),
    install_config: {
      version: {
        id: 'icv-0',
        created_at: new Date(Date.now() - day * 14).toISOString(),
        install_config_sync_id: 'ics-0',
        install_id: 'inst-1',
        install_name: 'acme-prod',
        file_path: 'installs/acme-prod.toml',
        created: true,
        status: { status: 'success' },
      },
    },
  },
]

export const Default = () => (
  <InstallUpdatesTimeline
    updates={mockUpdates}
    orgId="org-1"
    installId="inst-1"
    appId="app-1"
  />
)

export const Empty = () => <InstallUpdatesTimeline updates={[]} />

export const Loading = () => <InstallUpdatesTimeline updates={[]} isLoading />

export const HasMore = () => (
  <InstallUpdatesTimeline
    updates={mockUpdates}
    orgId="org-1"
    installId="inst-1"
    appId="app-1"
    hasMore
    onLoadMore={() => {}}
  />
)

// `frontend` is impacted only through the graph — its own config did not change, so it
// has matching checksums and no build change. Checks the timeline status badge for a
// failed app config version at the same time.
export const ImpactedAndFailed = () => (
  <InstallUpdatesTimeline
    updates={[
      {
        id: 'upd-4',
        type: 'app_config',
        created_at: new Date(Date.now() - 3600000).toISOString(),
        workflow_id: 'wkf-4',
        app_config: {
          version: {
            id: 'iacv-4',
            new_app_config_id: 'appcfg-4',
            status: {
              status: 'error',
              status_human_description:
                'stack version generation timed out after 30m',
            },
          },
          diff: {
            added: [],
            removed: [],
            unchanged: [],
            stack_changed: true,
            stack_impacts: ['permissions'],
            stack_impact_reasons: [
              { from: 'policy.maintenance-readonly', edge: 'stack_render' },
            ],
            changed: [
              {
                component_id: 'cmp-frontend',
                component_name: 'frontend',
                component_type: 'helm_chart',
                old_checksum: 'sha256:aaaa',
                new_checksum: 'sha256:aaaa',
                build_changed: false,
                impact_reasons: [
                  { from: 'component.api', edge: 'component_ref' },
                ],
              },
            ],
          },
        },
      },
    ]}
    orgId="org-1"
    installId="inst-1"
    appId="app-1"
  />
)
