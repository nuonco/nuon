'use client'

import { useContext } from 'react'
import { PaginationContext } from '@/providers/pagination-provider'

export function usePagination() {
  const context = useContext(PaginationContext)
  if (!context) {
    throw new Error('usePagination must be used within a PaginationProvider')
  }
  return context
}
