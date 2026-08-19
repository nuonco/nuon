import { type FC, useEffect, useRef } from 'react'
import posthog from 'posthog-js'
import { useAuth } from '@/hooks/use-auth'
import { useOrg } from '@/hooks/use-org'

let isPostHogInitialized = false

export const InitPostHog: FC<{ apiKey: string; apiHost?: string }> = ({
  apiKey,
  apiHost,
}) => {
  const { user, isLoading } = useAuth()
  const identifiedIdRef = useRef<string | undefined>(undefined)

  useEffect(() => {
    if (isPostHogInitialized || !apiKey) return
    posthog.init(apiKey, {
      api_host: apiHost || 'https://us.i.posthog.com',
      defaults: '2025-05-24',
      person_profiles: 'identified_only',
    })
    isPostHogInitialized = true
  }, [apiKey, apiHost])

  useEffect(() => {
    if (!isPostHogInitialized || isLoading || !user?.sub) return
    if (identifiedIdRef.current === user.sub) return
    identifiedIdRef.current = user.sub
    posthog.identify(user.sub, { email: user.email, name: user.name })
  }, [user, isLoading])

  return null
}

export const PostHogGroupOrg: FC = () => {
  const { org } = useOrg()
  const groupedIdRef = useRef<string | undefined>(undefined)

  useEffect(() => {
    if (!isPostHogInitialized || !org?.id || groupedIdRef.current === org.id) {
      return
    }
    groupedIdRef.current = org.id
    posthog.group('org', org.id, { name: org.name })
  }, [org?.id, org?.name])

  return null
}

export function capturePostHogEvent(
  event: string,
  properties?: Record<string, unknown>
) {
  if (!isPostHogInitialized) return
  posthog.capture(event, properties)
}
