export default {
  title: 'Install Components/QuickComponentManagementDropdown',
}

import { QuickComponentManagementDropdown } from './QuickComponentManagementDropdown'
import type { TInstallComponent } from '@/types'

const mockInstallComponent: TInstallComponent = {
  id: 'ic-1',
  component: {
    id: 'comp-1',
    name: 'api',
    type: 'helm_chart',
  },
} as TInstallComponent

export const Default = () => (
  <QuickComponentManagementDropdown
    installComponent={mockInstallComponent}
    orgId="org-1"
    installId="install-1"
  />
)
