export default {
  title: 'InstallResources/InstallResourceDetailPanel',
}

import { PanelStory } from '@/components/__stories__/helpers'
import type { TInstallResource } from '@/types'
import {
  InstallResourceDetailPanel,
  InstallResourceDetailPanelButton,
} from './InstallResourceDetailPanel'

const mockResource: TInstallResource = {
  org_id: 'org-1',
  install_id: 'install-1',
  install_component_id: 'instcmp-1',
  component_id: 'component-1',
  runner_id: 'runner-1',
  provider: 'kubernetes',
  api_group: 'apps',
  kind: 'Deployment',
  namespace: 'default',
  name: 'web-app',
  health: 'degraded',
  message: 'Waiting for 2 of 3 replicas to become ready.',
  native_status: 'Progressing',
  details: JSON.stringify({ replicas: 3, availableReplicas: 1 }),
  observed_at: new Date().toISOString(),
}

export const Default = () => (
  <div className="p-4">
    <InstallResourceDetailPanelButton onOpen={() => {}} />
  </div>
)

export const Panel = () => (
  <PanelStory label="Open resource details">
    <InstallResourceDetailPanel installResource={mockResource} />
  </PanelStory>
)
