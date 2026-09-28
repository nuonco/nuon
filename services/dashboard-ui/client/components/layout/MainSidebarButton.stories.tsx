export default {
  title: 'UI / Layout / Main sidebar button',
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
