import { PanelStory } from '@/components/__stories__/helpers'
import { InstallUpdateDetails } from './InstallUpdateDetails'
import type { TInstallUpdate } from '@/types'

export default {
  title: 'InstallUpdates/InstallUpdateDetails',
}

const mockUpdate: TInstallUpdate = {
  id: 'upd-3',
  created_at: new Date(Date.now() - 3600000).toISOString(),
  type: 'app_config',
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
      added: [
        {
          component_id: 'cmp-worker',
          component_name: 'worker',
          component_type: 'external_image',
        },
      ],
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
  workflow_id: 'wkf-3',
}

export const Default = () => (
  <PanelStory>
    <InstallUpdateDetails
      update={mockUpdate}
      orgId="org-1"
      installId="inst-1"
      appId="app-1"
    />
  </PanelStory>
)

export const NoImpacts = () => (
  <PanelStory>
    <InstallUpdateDetails
      update={{
        id: 'upd-1',
        type: 'inputs',
        created_at: new Date(Date.now() - 86400000).toISOString(),
        inputs: { input_config_id: 'inpcfg-1', keys: ['region'] },
      }}
      orgId="org-1"
      installId="inst-1"
      appId="app-1"
    />
  </PanelStory>
)

// The graph case: `frontend`'s own config is byte-identical, but an upstream role change
// reached it, so ComputeInstallConfigDiff promotes it out of `unchanged` into `changed`
// carrying only impact reasons. It should read as impacted, not as edited.
export const ImpactedButUnchanged = () => (
  <PanelStory>
    <InstallUpdateDetails
      update={{
        id: 'upd-4',
        type: 'app_config',
        created_at: new Date(Date.now() - 1800000).toISOString(),
        app_config: {
          version: {
            id: 'iacv-4',
            new_app_config_id: 'appcfg-4',
            status: { status: 'success' },
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
                  { from: 'role.maintenance', edge: 'operation_role' },
                ],
              },
            ],
          },
        },
      }}
      orgId="org-1"
      installId="inst-1"
      appId="app-1"
    />
  </PanelStory>
)

export const InstallConfig = () => (
  <PanelStory>
    <InstallUpdateDetails
      update={{
        id: 'upd-5',
        type: 'install_config',
        created_at: new Date(Date.now() - 7200000).toISOString(),
        install_config: {
          version: {
            id: 'icv-5',
            created_at: new Date(Date.now() - 7200000).toISOString(),
            install_config_sync_id: 'ics-5',
            install_id: 'inst-1',
            install_name: 'acme-prod',
            file_path: 'installs/acme-prod.toml',
            created: false,
            status: { status: 'success' },
          },
        },
      }}
      orgId="org-1"
      installId="inst-1"
      appId="app-1"
    />
  </PanelStory>
)

export const Failed = () => (
  <PanelStory>
    <InstallUpdateDetails
      update={{
        ...mockUpdate,
        id: 'upd-6',
        app_config: {
          ...mockUpdate.app_config!,
          version: {
            ...mockUpdate.app_config!.version,
            status: {
              status: 'error',
              status_human_description:
                'stack version generation timed out after 30m',
            },
          },
        },
      }}
      orgId="org-1"
      installId="inst-1"
      appId="app-1"
    />
  </PanelStory>
)

export const LongContent = () => (
  <PanelStory>
    <InstallUpdateDetails
      update={{
        ...mockUpdate,
        id: 'upd-7',
        app_config: {
          version: {
            ...mockUpdate.app_config!.version,
            app_branch_run: {
              ...mockUpdate.app_config!.version.app_branch_run,
              vcs_connection_commit: {
                sha: 'f00ba4c0ffee1234',
                message:
                  'Rotate the maintenance role and tighten its inline policy\n\nThe maintenance role previously carried a wildcard on s3, which the stack\nrenders into the install role. Narrow it to the bucket prefix and rotate the\nattached policy so every downstream component re-renders.',
                author_name: 'Jane Doe',
              },
            },
          },
          diff: {
            added: [],
            removed: [],
            unchanged: [],
            stack_changed: true,
            stack_impacts: [
              'permissions',
              'stack_config',
              'secrets',
              'runner_config',
            ],
            stack_impact_reasons: [
              { from: 'role.maintenance', edge: 'stack_render' },
              { from: 'policy.maintenance-readonly', edge: 'stack_render' },
              { from: 'secret.registry-token', edge: 'stack_render' },
            ],
            changed: Array.from({ length: 8 }, (_, index) => ({
              component_id: `cmp-${index}`,
              component_name: `some-fairly-long-component-name-${index}`,
              component_type: index % 2 ? 'helm_chart' : 'terraform_module',
              build_changed: index % 3 === 0,
              impact_reasons: [
                { from: 'role.maintenance', edge: 'operation_role' },
                { from: 'stack', edge: 'install_stack_outputs' },
              ],
            })),
          },
        },
      }}
      orgId="org-1"
      installId="inst-1"
      appId="app-1"
    />
  </PanelStory>
)
