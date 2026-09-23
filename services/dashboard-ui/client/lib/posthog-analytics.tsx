import { useEffect, useRef } from 'react'
import { useLocation } from 'react-router'
import posthog from 'posthog-js'
import { useAuth } from '@/hooks/use-auth'
import { useConfig } from '@/hooks/use-config'
import { useOrg } from '@/hooks/use-org'
import { useApp } from '@/hooks/use-app'
import { useInstall } from '@/hooks/use-install'
import type { IUser } from '@/types'

let initialized = false

const snakeCase = (key: string) =>
  key
    .replace(/([a-z0-9])([A-Z])/g, '$1_$2')
    .replace(/[\s-]+/g, '_')
    .toLowerCase()

const adaptProps = (props: Record<string, unknown>) =>
  Object.fromEntries(
    Object.entries(props).map(([key, value]) => [snakeCase(key), value])
  )

// Repeated reloads of one page within a short window signal a user waiting
// on progress the page is not showing. The streak counter lives in
// sessionStorage because a reload tears down all JS state.
const RELOAD_WINDOW_MS = 2 * 60 * 1000
const RELOAD_STREAK_KEY = 'nuon_reload_streak'

type ReloadStreak = { path: string; count: number; startedAt: number }

const readReloadStreak = (): ReloadStreak | null => {
  try {
    const raw = sessionStorage.getItem(RELOAD_STREAK_KEY)
    return raw ? (JSON.parse(raw) as ReloadStreak) : null
  } catch {
    return null
  }
}

const writeReloadStreak = (streak: ReloadStreak | null) => {
  try {
    if (streak) {
      sessionStorage.setItem(RELOAD_STREAK_KEY, JSON.stringify(streak))
    } else {
      sessionStorage.removeItem(RELOAD_STREAK_KEY)
    }
  } catch {
    // storage unavailable; streaks just won't span reloads
  }
}

// True only for the first pageview of this document lifetime; a reload
// re-runs the bundle, so the flag resets exactly when it should.
let documentPageviewSeen = false

const isDocumentReload = () => {
  const nav = performance.getEntriesByType('navigation')[0] as
    | PerformanceNavigationTiming
    | undefined
  return nav?.type === 'reload'
}

// Pogo-sticking: A -> B -> A within seconds, a failed expectation on B.
const POGO_WINDOW_MS = 10_000
const navTrail: { path: string; at: number }[] = []

// Stale-tab return: back after 5+ minutes away, and bailing out again
// within 30s — the user did not trust what they saw. The marker lives in
// sessionStorage so a bounce-by-reload (fresh bundle) still counts.
const STALE_AFTER_MS = 5 * 60_000
const STALE_BOUNCE_WINDOW_MS = 30_000
const STALE_RETURN_KEY = 'nuon_stale_return'

type StaleReturn = { at: number; minutesGone: number }

const readStaleReturn = (): StaleReturn | null => {
  try {
    const raw = sessionStorage.getItem(STALE_RETURN_KEY)
    return raw ? (JSON.parse(raw) as StaleReturn) : null
  } catch {
    return null
  }
}

const writeStaleReturn = (stale: StaleReturn | null) => {
  try {
    if (stale) {
      sessionStorage.setItem(STALE_RETURN_KEY, JSON.stringify(stale))
    } else {
      sessionStorage.removeItem(STALE_RETURN_KEY)
    }
  } catch {
    // storage unavailable; bounce-by-reload just won't be counted
  }
}

export const InitPostHog = ({ apiKey }: { apiKey: string }) => {
  const { user, isLoading } = useAuth()
  const { isByoc, byocName, version, posthogReplayEnabled } = useConfig()
  const { pathname } = useLocation()
  const identifiedSubRef = useRef<string | null>(null)

  useEffect(() => {
    if (!apiKey || initialized) return
    posthog.init(apiKey, {
      api_host: '/ingest',
      ui_host: 'https://us.posthog.com',
      autocapture: true,
      capture_pageview: false,
      disable_session_recording: !posthogReplayEnabled,
    })
    posthog.register({
      deployment: isByoc ? 'byoc' : 'cloud',
      ...(byocName ? { byoc_name: byocName } : {}),
      ...(version ? { nuon_version: version } : {}),
    })
    initialized = true
  }, [apiKey, isByoc, byocName, version, posthogReplayEnabled])

  useEffect(() => {
    if (!initialized || isLoading) return
    if (!user?.sub) {
      if (identifiedSubRef.current) {
        posthog.reset()
        identifiedSubRef.current = null
      }
      return
    }
    posthog.identify(user.email || user.sub, {
      email: user.email,
      name: user.name,
      account_id: user.sub,
    })
    identifiedSubRef.current = user.sub
  }, [user, isLoading])

  useEffect(() => {
    if (!initialized) return
    const firstPageview = !documentPageviewSeen
    documentPageviewSeen = true
    const now = Date.now()

    // Stale-tab bounce, via reload (handled in the firstPageview branch
    // below) or via navigating away right after coming back.
    const stale = readStaleReturn()
    if (stale) {
      writeStaleReturn(null)
      if (now - stale.at <= STALE_BOUNCE_WINDOW_MS) {
        posthog.capture('stale_tab_bounce', {
          path: pathname,
          minutes_gone: stale.minutesGone,
          via: firstPageview ? 'reload' : 'navigate',
        })
      }
    }

    if (firstPageview && isDocumentReload()) {
      const prev = readReloadStreak()
      const streak =
        prev && prev.path === pathname && now - prev.startedAt <= RELOAD_WINDOW_MS
          ? { path: pathname, count: prev.count + 1, startedAt: prev.startedAt }
          : { path: pathname, count: 1, startedAt: now }
      writeReloadStreak(streak)
      posthog.capture('$pageview', {
        is_reload: true,
        reload_count: streak.count,
      })
      if (streak.count >= 2) {
        posthog.capture('rage_reload', {
          path: pathname,
          reload_count: streak.count,
          seconds_since_first: Math.round((now - streak.startedAt) / 1000),
        })
      }
      return
    }

    if (!firstPageview) {
      const awayFrom = navTrail.length === 2 ? navTrail[0] : null
      if (
        awayFrom &&
        awayFrom.path === pathname &&
        now - awayFrom.at <= POGO_WINDOW_MS
      ) {
        posthog.capture('pogo_stick', {
          path: pathname,
          away_path: navTrail[navTrail.length - 1].path,
          seconds_away: Math.round((now - awayFrom.at) / 1000),
        })
      }
    }
    navTrail.push({ path: pathname, at: now })
    if (navTrail.length > 2) navTrail.shift()

    writeReloadStreak(null)
    posthog.capture('$pageview', { is_reload: false })
  }, [pathname])

  useEffect(() => {
    if (!initialized) return
    let hiddenAt: number | null = null
    const onVisibility = () => {
      if (document.visibilityState === 'hidden') {
        hiddenAt = Date.now()
        return
      }
      if (hiddenAt === null) return
      const goneMs = Date.now() - hiddenAt
      hiddenAt = null
      if (goneMs < STALE_AFTER_MS) return
      const minutesGone = Math.round(goneMs / 60_000)
      writeStaleReturn({ at: Date.now(), minutesGone })
      posthog.capture('stale_tab_return', { minutes_gone: minutesGone })
    }
    document.addEventListener('visibilitychange', onVisibility)
    return () =>
      document.removeEventListener('visibilitychange', onVisibility)
  }, [])

  return null
}

export const PostHogOrgProperties = () => {
  const { org } = useOrg()

  useEffect(() => {
    if (!initialized || !org?.id) return
    posthog.register({ org_id: org.id })
    posthog.group('organization', org.id, { name: org.name })
  }, [org?.id, org?.name])

  return null
}

export const PostHogAppProperties = () => {
  const { app } = useApp()

  useEffect(() => {
    if (!initialized || !app?.id) return
    posthog.register({ app_id: app.id })
    return () => posthog.unregister('app_id')
  }, [app?.id])

  return null
}

export const PostHogInstallProperties = () => {
  const { install } = useInstall()

  useEffect(() => {
    if (!initialized || !install?.id) return
    posthog.register({ install_id: install.id })
    return () => posthog.unregister('install_id')
  }, [install?.id])

  return null
}

interface ITrackEvent {
  event: string
  props?: Record<string, unknown>
  status: 'ok' | 'error'
  user: IUser
}

export function trackEvent({ event, status, props = {} }: ITrackEvent) {
  if (!initialized) return
  posthog.capture(event, { status, ...adaptProps(props) })
}
