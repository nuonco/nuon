export default {
  title: 'Features / Installs / Settings panel',
}

import { useState } from 'react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useConfig } from '@/hooks/use-config'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { ConfigContext } from '@/providers/config-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { Panel } from '@/components/surfaces/Panel'
import { InstallSettingsPanelContent } from './InstallSettingsPanelContent'

const Example = ({
  isByoc = true,
  relayConfigured = true,
  hasStackEndpoint = false,
}) => {
  const config = useConfig()
  const { org } = useOrg()
  const { install } = useInstall()
  const [client] = useState(() => {
    const client = new QueryClient({
      defaultOptions: { queries: { retry: false, staleTime: Infinity } },
    })
    client.setQueryData(['install-stack', org.id, install.id], {
      install_stack_outputs: {
        data_contents: {
          telemetry_endpoint: hasStackEndpoint ? 'http://10.0.0.4:4318' : '',
        },
      },
    })
    client.setQueryData(['install-telemetry', org.id, install.id], {
      enabled: false,
      override: null,
      org_default: false,
      relay_configured: relayConfigured,
    })
    return client
  })

  return (
    <ConfigContext.Provider value={{ ...config, isByoc, isDev: false }}>
      <QueryClientProvider client={client}>
        <SurfacesProvider>
          <div className="relative h-screen w-full">
            <Panel heading="Settings" isVisible size="half">
              <InstallSettingsPanelContent />
            </Panel>
          </div>
        </SurfacesProvider>
      </QueryClientProvider>
    </ConfigContext.Provider>
  )
}

export const Default = () => <Example />
export const CloudWithRelay = () => <Example isByoc={false} hasStackEndpoint />
export const CloudWithoutRelay = () => (
  <Example isByoc={false} relayConfigured={false} hasStackEndpoint />
)
export const CloudMissingStackEndpoint = () => <Example isByoc={false} />
