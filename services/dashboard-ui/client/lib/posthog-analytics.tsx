import { useEffect } from 'react'
import { useLocation } from 'react-router'
import posthog from 'posthog-js'
import { useAuth } from '@/hooks/use-auth'
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

export const InitPostHog = ({ apiKey }: { apiKey: string }) => {
  const { user, isLoading } = useAuth()
  const { pathname } = useLocation()

  useEffect(() => {
    if (!apiKey || initialized) return
    posthog.init(apiKey, {
      api_host: '/ingest',
      ui_host: 'https://us.posthog.com',
      autocapture: true,
      capture_pageview: false,
    })
    initialized = true
  }, [apiKey])

  useEffect(() => {
    if (!initialized || isLoading || !user?.sub) return
    posthog.identify(user.sub, { email: user.email, name: user.name })
  }, [user, isLoading])

  useEffect(() => {
    if (!initialized) return
    posthog.capture('$pageview')
  }, [pathname])

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
