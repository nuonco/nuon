export default {
  title: 'InstallHealth/RefreshClusterAccess',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { OrgContext } from '@/providers/org-provider'
import type { TOrg } from '@/types'
import { RefreshClusterAccessModal } from './RefreshClusterAccess'

// The modal renders the RoleSelector container, which reads org.id on render.
const mockOrg = {
  org: { id: 'orgstory', name: 'Story org' } as TOrg,
  refresh: () => {},
}

export const Default = () => (
  <OrgContext.Provider value={mockOrg}>
    <ModalStory>
      <RefreshClusterAccessModal installId="inl123" />
    </ModalStory>
  </OrgContext.Provider>
)
