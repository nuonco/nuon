'use client'

import { usePathname } from 'next/navigation'
import { createContext, useState, type ReactNode } from 'react'
import type { TNavLink } from '@/types'

type BreadcrumbContextValue = {
  breadcrumbLinks: TNavLink[]
  isLoading: boolean
  updateBreadcrumb: (links: TNavLink[]) => void
}

export const BreadcrumbContext = createContext<
  BreadcrumbContextValue | undefined
>(undefined)

export function BreadcrumbProvider({ children }: { children: ReactNode }) {
  const pathname = usePathname()
  const segments = pathname.split('/').filter(Boolean)
  const [breadcrumbLinks, setBreadcrumbLinks] = useState<TNavLink[]>(
    segments?.map((s) => ({ path: `/${s}`, text: s }))
  )
  const [isLoading, setIsLoading] = useState<boolean>(true)

  const updateBreadcrumb = (links: TNavLink[]) => {
    setIsLoading(true)
    setBreadcrumbLinks(links)
    setIsLoading(false)
  }

  return (
    <BreadcrumbContext.Provider
      value={{
        breadcrumbLinks,
        isLoading,
        updateBreadcrumb,
      }}
    >
      {children}
    </BreadcrumbContext.Provider>
  )
}
