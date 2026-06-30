export default {
  title: 'Apps/Config/AppKubernetesContexts',
}

import { AppKubernetesContexts } from './AppKubernetesContexts'

export const Default = () => (
  <AppKubernetesContexts
    appConfig={
      {
        kubernetes_contexts: {
          contexts: [
            {
              id: 'apc1',
              name: 'secondary',
              source_component_name: 'cluster',
              source_component_id: 'cmp1',
              org_id: 'org1',
              app_id: 'app1',
            },
          ],
        },
      } as any
    }
  />
)

export const Multiple = () => (
  <AppKubernetesContexts
    appConfig={
      {
        kubernetes_contexts: {
          contexts: [
            {
              id: 'apc1',
              name: 'data-cluster',
              source_component_name: 'data-eks',
              source_component_id: 'cmp1',
              org_id: 'org1',
              app_id: 'app1',
            },
            {
              id: 'apc2',
              name: 'shared-prod',
              source_component_name: 'shared-eks',
              source_component_id: 'cmp2',
              org_id: 'org1',
              app_id: 'app1',
            },
          ],
        },
      } as any
    }
  />
)

export const Empty = () => (
  <AppKubernetesContexts
    appConfig={
      {
        kubernetes_contexts: {
          contexts: [],
        },
      } as any
    }
  />
)
