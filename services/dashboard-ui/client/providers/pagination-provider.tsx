'use client'

import { createContext, useState, useMemo, type ReactNode } from 'react'

type TPaginationContext = {
  isPaginating: boolean
  setIsPaginating: (isPaginating: boolean) => void
}

export const PaginationContext = createContext<TPaginationContext | undefined>(
  undefined
)

export function PaginationProvider({ children }: { children: ReactNode }) {
  const [isPaginating, setIsPaginating] = useState(false)

  const value = useMemo(
    () => ({
      isPaginating,
      setIsPaginating,
    }),
    [isPaginating]
  )

  return (
    <PaginationContext.Provider value={value}>
      {children}
    </PaginationContext.Provider>
  )
}
