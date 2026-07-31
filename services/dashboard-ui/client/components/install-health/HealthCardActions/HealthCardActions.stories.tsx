export default {
  title: 'InstallHealth/HealthCardActions',
}

import { OrgContext } from '@/providers/org-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import type { TOrg } from '@/types'
import { HealthCardActions } from './HealthCardActions'

const mockOrg = {
  org: { id: 'orgstory', name: 'Story org' } as TOrg,
  refresh: () => {},
}

export const Default = () => (
  <OrgContext.Provider value={mockOrg}>
    <SurfacesProvider>
      <div className="flex justify-end p-8">
        <HealthCardActions installId="inl123" />
      </div>
    </SurfacesProvider>
  </OrgContext.Provider>
)
