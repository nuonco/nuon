'use client'

import { useContext } from 'react'
import { SurfacesContext } from '@/contexts/surfaces-context'

export function useSurfaces() {
  const ctx = useContext(SurfacesContext)
  if (!ctx)
    throw new Error('useSurfaces must be used within a SurfacesProvider')
  return ctx
}
