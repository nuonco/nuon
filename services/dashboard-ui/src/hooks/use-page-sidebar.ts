import { useContext } from 'react'
import { PageSidebarContext } from '@/providers/page-sidebar-provider'

export function usePageSidebar() {
  const ctx = useContext(PageSidebarContext)
  if (!ctx) {
    throw new Error('usePageSidebar must be used within an PageSidebarProvider')
  }
  return ctx
}
