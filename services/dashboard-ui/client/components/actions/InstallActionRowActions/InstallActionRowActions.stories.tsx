export default {
  title: 'Features / Installs / Actions / Row actions',
}

import type { TAction, TActionConfig } from '@/types'
import { InstallActionRowActions } from './InstallActionRowActions'

const action = {
  id: 'act-1',
  name: 'rotate-secrets',
} as TAction

const config = {
  id: 'acfg-1',
  action_workflow_id: 'act-1',
  triggers: [{ id: 'trg-1', type: 'manual' }],
} as TActionConfig

export const Runnable = () => (
  <InstallActionRowActions
    action={action}
    actionId="act-1"
    config={config}
    name="rotate-secrets"
  />
)

export const NoManualTrigger = () => (
  <InstallActionRowActions
    action={action}
    actionId="act-1"
    config={{ ...config, triggers: [{ id: 'trg-2', type: 'cron' }] }}
    name="rotate-secrets"
  />
)

export const Removed = () => (
  <InstallActionRowActions
    action={action}
    actionId="act-1"
    config={config}
    name="rotate-secrets"
    removed
  />
)
