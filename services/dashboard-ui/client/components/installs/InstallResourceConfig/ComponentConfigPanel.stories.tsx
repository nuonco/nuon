import { PanelStory } from '@/components/__stories__/helpers'
import type { TComponentConfig } from '@/types'
import { ComponentConfigDetails } from './ComponentConfigPanel'

export default {
  title: 'Features / Installs / Component config',
}

const config = {
  id: 'ccc-1',
  component_id: 'cmp-1',
  app_config_id: 'cfg-v2',
  type: 'helm_chart',
  version: 4,
  build_timeout: '30m',
  deploy_timeout: '1h',
  helm: {
    chart_name: 'acme-api',
    namespace: 'payments',
  },
} as TComponentConfig

const run = {
  id: 'run-1',
  status: 'succeeded',
  created_at: '2026-09-18T15:04:00Z',
  vcs_connection_commit: {
    sha: 'a1b2c3d4e5f6a7b8',
    message: 'Pin acme-api chart to 2.4.0',
    author_name: 'Ada Lovelace',
    created_at: '2026-09-18T15:02:00Z',
  },
}

export const Config = () => (
  <PanelStory label="Open config">
    <ComponentConfigDetails
      config={config}
      emptyMessage="This component is not in the app config this install is using."
      run={run}
      runHref="#"
    />
  </PanelStory>
)

export const Behind = () => (
  <PanelStory label="Open config">
    <ComponentConfigDetails
      behind
      config={config}
      emptyMessage="This component is not in the app config this install is using."
      run={run}
      runHref="#"
    />
  </PanelStory>
)

export const Missing = () => (
  <PanelStory label="Open config">
    <ComponentConfigDetails emptyMessage="This component is not in the app config this install is using." />
  </PanelStory>
)
