import { useContext } from 'react'
import { ConsentContext } from '@/providers/consent-provider'

export function useConsent() {
  const context = useContext(ConsentContext)
  if (!context) {
    throw new Error('useConsent must be used within a ConsentProvider')
  }
  return context
}
