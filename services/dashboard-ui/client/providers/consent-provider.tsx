import { createContext, useCallback, useEffect, useState, type ReactNode } from 'react'
import { DateTime } from 'luxon'

export type TConsentStatus = 'unknown' | 'granted' | 'denied'

interface ConsentSettings {
  status: 'granted' | 'denied'
  decidedAt: string
}

interface ConsentContextType {
  consent: TConsentStatus
  accept: () => void
  decline: () => void
}

export const ConsentContext = createContext<ConsentContextType | null>(null)

const STORAGE_KEY = 'analytics_consent'

export function ConsentProvider({ children }: { children: ReactNode }) {
  const [consent, setConsent] = useState<TConsentStatus>('unknown')

  useEffect(() => {
    try {
      const stored = localStorage.getItem(STORAGE_KEY)
      if (!stored) return
      const parsed = JSON.parse(stored) as ConsentSettings
      if (parsed.status === 'granted' || parsed.status === 'denied') {
        setConsent(parsed.status)
      }
    } catch (error) {
      console.warn('Failed to load analytics consent from localStorage:', error)
    }
  }, [])

  const persist = useCallback((status: 'granted' | 'denied') => {
    try {
      const settings: ConsentSettings = {
        status,
        decidedAt: DateTime.now().toISO() ?? '',
      }
      localStorage.setItem(STORAGE_KEY, JSON.stringify(settings))
    } catch (error) {
      console.warn('Failed to save analytics consent to localStorage:', error)
    }
    setConsent(status)
  }, [])

  const accept = useCallback(() => persist('granted'), [persist])
  const decline = useCallback(() => persist('denied'), [persist])

  return (
    <ConsentContext.Provider value={{ consent, accept, decline }}>
      {children}
    </ConsentContext.Provider>
  )
}
