export default {
  title: 'Installs/InstallSettingsPanel',
}

import { SurfacesProvider } from '@/providers/surfaces-provider'
import { Panel } from '@/components/surfaces/Panel'

export const Default = () => (
  <SurfacesProvider>
    <div className="relative h-screen w-full">
      <Panel heading="Settings" isVisible size="default" />
    </div>
  </SurfacesProvider>
)
