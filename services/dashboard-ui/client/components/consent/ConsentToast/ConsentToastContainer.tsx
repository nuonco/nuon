import { useEffect, useRef } from 'react'
import { useConfig } from '@/hooks/use-config'
import { useConsent } from '@/hooks/use-consent'
import { useToast } from '@/hooks/use-toast'
import { ConsentToast } from './ConsentToast'

export const ConsentToastContainer = () => {
  const { posthogKey } = useConfig()
  const { consent, accept, decline } = useConsent()
  const { addToast } = useToast()
  const added = useRef(false)

  useEffect(() => {
    if (!posthogKey || consent !== 'unknown' || added.current) return
    added.current = true
    addToast(<ConsentToast onAccept={accept} onDecline={decline} />)
  }, [posthogKey, consent, accept, decline, addToast])

  return null
}
