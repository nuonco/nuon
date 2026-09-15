export default {
  title: 'Installs/InstallSettingsPanel',
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

export const Default = () => {
  const config = useConfig()
  const { org } = useOrg()
  const { install } = useInstall()
  const [client] = useState(() => {
    const client = new QueryClient({
      defaultOptions: { queries: { retry: false, staleTime: Infinity } },
    })
    client.setQueryData(['install-stack', org.id, install.id], {
      install_stack_outputs: { data_contents: {} },
    })
    client.setQueryData(['install-telemetry', org.id, install.id], {
      enabled: false,
    })
    return client
  })

  return (
    <ConfigContext.Provider value={{ ...config, isByoc: true }}>
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
