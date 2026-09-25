export const MAX_ACTIVITY_PINS = 4

export type TActivityPin = {
  kind: 'action' | 'runbook'
  id: string
}

const empty: TActivityPin[] = []

const snapshots = new Map<string, { raw: string | null; pins: TActivityPin[] }>()
const listeners = new Set<() => void>()

const isPin = (value: unknown): value is TActivityPin => {
  if (!value || typeof value !== 'object') return false
  const pin = value as TActivityPin
  return (
    (pin.kind === 'action' || pin.kind === 'runbook') &&
    typeof pin.id === 'string' &&
    pin.id.length > 0
  )
}

const parsePins = (raw: string | null): TActivityPin[] => {
  if (!raw) return empty
  try {
    const parsed = JSON.parse(raw)
    if (!Array.isArray(parsed)) return empty
    const pins = parsed.filter(isPin).slice(0, MAX_ACTIVITY_PINS)
    return pins.length ? pins : empty
  } catch {
    return empty
  }
}

export const activityPinStorageKey = (userId: string, installId: string) =>
  `nuon.activity-pins:${userId}:${installId}`

export const readActivityPins = (key: string): TActivityPin[] => {
  let raw: string | null = null
  try {
    raw = localStorage.getItem(key)
  } catch {
    raw = null
  }
  const cached = snapshots.get(key)
  if (cached && cached.raw === raw) return cached.pins
  const pins = parsePins(raw)
  snapshots.set(key, { raw, pins })
  return pins
}

export const writeActivityPins = (key: string, pins: TActivityPin[]) => {
  const next = pins.slice(0, MAX_ACTIVITY_PINS)
  const raw = JSON.stringify(next)
  try {
    localStorage.setItem(key, raw)
  } catch {
    return
  }
  snapshots.set(key, { raw, pins: next.length ? next : empty })
  listeners.forEach((listener) => listener())
}

export const toggleActivityPin = (key: string, pin: TActivityPin) => {
  const current = readActivityPins(key)
  const exists = current.some(
    (item) => item.kind === pin.kind && item.id === pin.id
  )
  if (exists) {
    writeActivityPins(
      key,
      current.filter((item) => !(item.kind === pin.kind && item.id === pin.id))
    )
    return
  }
  if (current.length >= MAX_ACTIVITY_PINS) return
  writeActivityPins(key, [...current, pin])
}

export const subscribeActivityPins = (listener: () => void) => {
  listeners.add(listener)
  const onStorage = (event: StorageEvent) => {
    if (event.key && !event.key.startsWith('nuon.activity-pins:')) return
    if (event.key) snapshots.delete(event.key)
    listener()
  }
  window.addEventListener('storage', onStorage)
  return () => {
    listeners.delete(listener)
    window.removeEventListener('storage', onStorage)
  }
}
