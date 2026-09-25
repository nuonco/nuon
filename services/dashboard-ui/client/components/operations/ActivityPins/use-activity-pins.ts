import { useSyncExternalStore } from 'react'
import { useAuth } from '@/hooks/use-auth'
import { useInstall } from '@/hooks/use-install'
import {
  activityPinStorageKey,
  readActivityPins,
  subscribeActivityPins,
  toggleActivityPin,
  MAX_ACTIVITY_PINS,
  type TActivityPin,
} from './storage'

const empty: TActivityPin[] = []

export const useActivityPins = () => {
  const { user } = useAuth()
  const { install } = useInstall()
  const userId = user?.sub || user?.email || ''
  const installId = install?.id || ''
  const key = userId && installId ? activityPinStorageKey(userId, installId) : ''

  const pins = useSyncExternalStore(
    subscribeActivityPins,
    () => (key ? readActivityPins(key) : empty),
    () => empty
  )

  return {
    atLimit: pins.length >= MAX_ACTIVITY_PINS,
    isPinned: (pin: TActivityPin) =>
      pins.some((item) => item.kind === pin.kind && item.id === pin.id),
    pins,
    toggle: (pin: TActivityPin) => {
      if (!key) return
      toggleActivityPin(key, pin)
    },
  }
}
