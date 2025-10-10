'use client'

import { usePathname } from 'next/navigation'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { ToastProvider } from '@/providers/toast-provider'

export function isInstallsOrApps(pathname: string): boolean {
  const segments = pathname.replace(/^\/|\/$/g, '').split('/')

  return segments[1] === 'installs' || segments[1] === 'apps'
}

export default function Template({ children }: { children: React.ReactNode }) {
  const pathname = usePathname()

  return isInstallsOrApps(pathname) ? (
    children
  ) : (
    <ToastProvider>
      <SurfacesProvider>{children}</SurfacesProvider>
    </ToastProvider>
  )
}
