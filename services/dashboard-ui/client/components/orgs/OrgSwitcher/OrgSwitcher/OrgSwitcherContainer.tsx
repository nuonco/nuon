import { useOrg } from '@/hooks/use-org'
import { useSidebar } from '@/hooks/use-sidebar'
import { OrgSwitcher } from './OrgSwitcher'

export const OrgSwitcherContainer = () => {
  const { isSidebarOpen } = useSidebar()
  const { org } = useOrg()

  if (!org) return null

  return <OrgSwitcher org={org} isSidebarOpen={isSidebarOpen} />
}
