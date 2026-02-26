'use client'

import { useContext } from 'react'
import { SurfacesContext } from '@/providers/surfaces-provider'

export function useSurfaces() {
  const ctx = useContext(SurfacesContext)
  if (!ctx)
    throw new Error('useSurfaces must be used within a SurfacesProvider')
  return ctx
}
