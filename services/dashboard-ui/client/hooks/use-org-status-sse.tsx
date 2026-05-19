import { useContext } from 'react'
import { OrgStatusSSEContext } from '@/providers/org-status-sse-provider'

export function useOrgStatusSSE() {
  const ctx = useContext(OrgStatusSSEContext)
  return { sseConnected: ctx?.sseConnected ?? false }
}
