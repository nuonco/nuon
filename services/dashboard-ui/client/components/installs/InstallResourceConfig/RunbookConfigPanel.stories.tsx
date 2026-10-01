import { PanelStory } from '@/components/__stories__/helpers'
import type { TRunbookConfig } from '@/lib/ctl-api/apps/runbooks'
import { RunbookConfigDetails } from './RunbookConfigPanel'

export default {
  title: 'Features / Installs / Runbook config',
}

const config = {
  id: 'rbc-1',
  runbook_id: 'rbk-1',
  app_config_id: 'cfg-v2',
  inputs: [
    {
      id: 'in-1',
      name: 'region',
      display_name: 'Region',
      required: true,
      idx: 0,
    },
  ],
  steps: [
    {
      id: 'stp-1',
      idx: 0,
      name: 'Scale the worker',
      type: 'action',
      action_workflow_id: 'act-1',
    },
  ],
} as TRunbookConfig

export const Config = () => (
  <PanelStory label="Open config">
    <RunbookConfigDetails
      config={config}
      run={{
        id: 'run-1',
        status: 'succeeded',
        created_at: '2026-09-18T15:04:00Z',
        vcs_connection_commit: {
          sha: 'a1b2c3d4e5f6a7b8',
          message: 'Add checkout retry to the payments worker',
          author_name: 'Ada Lovelace',
          created_at: '2026-09-18T15:02:00Z',
        },
      }}
      runHref="#"
    />
  </PanelStory>
)

export const Missing = () => (
  <PanelStory label="Open config">
    <RunbookConfigDetails />
  </PanelStory>
)
