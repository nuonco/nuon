'use client'

import { useContext } from 'react'
import { SidebarContext } from '@/contexts/sidebar-context'

export function useSidebar() {
  const ctx = useContext(SidebarContext)
  if (!ctx) {
    throw new Error('useSidebar must be used within an SidebarProvider')
  }
  return ctx
}
