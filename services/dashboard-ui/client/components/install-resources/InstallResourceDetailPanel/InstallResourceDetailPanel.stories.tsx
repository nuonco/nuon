export default {
  title: 'Features / Installs / Resources / Detail panel',
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
  details: JSON.stringify({
    spec: { replicas: 3 },
    status: { replicas: 3, availableReplicas: 1, readyReplicas: 1 },
  }),
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

export const RouteParentConditions = () => (
  <PanelStory label="Open route conditions">
    <InstallResourceDetailPanel
      installResource={{
        provider: 'kubernetes',
        api_group: 'gateway.networking.k8s.io',
        kind: 'HTTPRoute',
        namespace: 'acme',
        name: 'api-route',
        observed_at: new Date().toISOString(),
        details: JSON.stringify({
          spec: { hostnames: ['api.example.com'] },
          status: {
            parents: [
              {
                parentRef: { name: 'public-gateway', sectionName: 'https' },
                controllerName: 'example.com/gateway',
                conditions: [
                  { type: 'Accepted', status: 'True', reason: 'Accepted' },
                  {
                    type: 'ResolvedRefs',
                    status: 'True',
                    reason: 'ResolvedRefs',
                  },
                ],
              },
              {
                parentRef: { name: 'internal-gateway', namespace: 'edge' },
                controllerName: 'example.com/internal-gateway',
                conditions: [
                  {
                    type: 'Accepted',
                    status: 'False',
                    reason: 'NotAllowedByListeners',
                    message: 'Route namespace is not allowed by this listener.',
                  },
                  { type: 'ResolvedRefs', status: 'Unknown' },
                ],
              },
            ],
          },
        }),
      }}
    />
  </PanelStory>
)

export const MissingDetails = () => (
  <PanelStory label="Open resource details">
    <InstallResourceDetailPanel
      installResource={{
        ...mockResource,
        details: undefined,
        native_status: undefined,
        runner_id: undefined,
        message: undefined,
      }}
    />
  </PanelStory>
)

export const UnstructuredCheck = () => (
  <PanelStory label="Open resource details">
    <InstallResourceDetailPanel
      installResource={{
        name: 'database-check',
        kind: 'CustomCheck',
        provider: 'custom',
        health: 'unhealthy',
        observed_at: new Date().toISOString(),
        details: 'Connection timed out [db:5432]. Retrying /ready.',
      }}
    />
  </PanelStory>
)
