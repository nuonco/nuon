import { PanelStory } from '@/components/__stories__/helpers'
import type { TActionConfig } from '@/types'
import { ActionConfigDetails } from './ActionConfigPanel'

export default {
  title: 'Features / Installs / Action config',
}

const config = {
  id: 'awc-1',
  action_workflow_id: 'act-1',
  app_config_id: 'cfg-v2',
  image: 'ghcr.io/acme/payments-worker:2.4.0',
  role: 'acme-action',
  timeout: 120_000_000_000,
  enable_kube_config: { bool: true, valid: true },
  triggers: [
    { id: 'trg-1', type: 'manual' },
    { id: 'trg-2', type: 'cron', cron_schedule: '0 */6 * * *' },
  ],
  steps: [
    {
      id: 'stp-1',
      idx: 0,
      name: 'Check queue depth',
      command: 'acme queue depth --warn-above 100',
    },
  ],
} as TActionConfig

export const Config = () => (
  <PanelStory label="Open config">
    <ActionConfigDetails
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
    <ActionConfigDetails />
  </PanelStory>
)
