import { useContext } from 'react'
import { OrgContext } from '@/providers/org-provider'
import type { TOrg } from '@/types'

export function useOrg(): { org: TOrg; refresh: () => void } {
  const ctx = useContext(OrgContext)
  if (!ctx) {
    throw new Error('useOrg must be used within an OrgProvider')
  }
  return ctx
}
