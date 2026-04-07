export default {
  title: 'Layout/MainSidebarButton',
}

import { SidebarProvider } from '@/providers/sidebar-provider'
import { MainSidebarButton } from './MainSidebarButton'

export const Default = () => (
  <SidebarProvider>
    <MainSidebarButton />
  </SidebarProvider>
)

export const Mobile = () => (
  <SidebarProvider>
    <MainSidebarButton variant="mobile" />
  </SidebarProvider>
)

export const MobileClose = () => (
  <SidebarProvider>
    <MainSidebarButton variant="mobile-close" />
  </SidebarProvider>
)
