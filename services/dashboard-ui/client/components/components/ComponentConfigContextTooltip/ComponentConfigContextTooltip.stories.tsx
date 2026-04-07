export default {
  title: 'Components/ComponentConfigContextTooltip',
}

import { SurfacesProvider } from '@/providers/surfaces-provider'
import { ComponentConfigContextTooltip } from './ComponentConfigContextTooltip'

export const Default = () => (
  <SurfacesProvider>
    <ComponentConfigContextTooltip
      config={{ component_id: 'comp-1', version: 3, type: 'terraform_module', terraform_module: { version: '1.5.0' } } as any}
      isLoading={false}
      hasError={false}
      orgId="org-1"
      appId="app-1"
      addModal={() => ''}
    />
  </SurfacesProvider>
)

export const Loading = () => (
  <SurfacesProvider>
    <ComponentConfigContextTooltip
      config={null}
      isLoading={true}
      hasError={false}
      orgId="org-1"
      appId="app-1"
      addModal={() => ''}
    />
  </SurfacesProvider>
)

export const Error = () => (
  <SurfacesProvider>
    <ComponentConfigContextTooltip
      config={null}
      isLoading={false}
      hasError={true}
      orgId="org-1"
      appId="app-1"
      addModal={() => ''}
    />
  </SurfacesProvider>
)
