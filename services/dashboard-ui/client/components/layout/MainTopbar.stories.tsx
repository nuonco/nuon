export default {
  title: 'Layout/MainTopbar',
}

import { SidebarProvider } from '@/providers/sidebar-provider'
import { MainTopbar } from './MainTopbar'
import { Text } from '@/components/common/Text'

export const Default = () => (
  <SidebarProvider>
    <MainTopbar>
      <Text variant="subtext" theme="neutral">Dashboard</Text>
    </MainTopbar>
  </SidebarProvider>
)

export const HideSidebarButtons = () => (
  <SidebarProvider>
    <MainTopbar hideSidebarButtons>
      <Text variant="subtext" theme="neutral">Single page layout</Text>
    </MainTopbar>
  </SidebarProvider>
)
