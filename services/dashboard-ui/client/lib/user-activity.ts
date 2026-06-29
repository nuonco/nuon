const ACTIVITY_EVENTS = [
  'pointerdown',
  'pointermove',
  'keydown',
  'wheel',
  'touchstart',
] as const

let lastActivityAt = Date.now()
let tracking = false

const markActivity = () => {
  lastActivityAt = Date.now()
}

export function ensureActivityTracking() {
  if (tracking || typeof window === 'undefined') return
  tracking = true
  for (const event of ACTIVITY_EVENTS) {
    window.addEventListener(event, markActivity, { passive: true })
  }
}

export function isRecentlyActive(windowMs: number) {
  return Date.now() - lastActivityAt < windowMs
}

export function onNextActivity(callback: () => void) {
  const handler = () => {
    cleanup()
    callback()
  }
  const cleanup = () => {
    for (const event of ACTIVITY_EVENTS) {
      window.removeEventListener(event, handler)
    }
  }
  for (const event of ACTIVITY_EVENTS) {
    window.addEventListener(event, handler, { passive: true })
  }
  return cleanup
}
