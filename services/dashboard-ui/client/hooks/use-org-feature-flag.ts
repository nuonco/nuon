import { useEffect } from 'react'
import { captureFeatureFlagEvaluation } from '@/lib/posthog-analytics'
import { useOrg } from '@/hooks/use-org'

const evaluated = new Set<string>()

/**
 * Reads an org feature flag and reports each evaluation to PostHog, so flag
 * usage, rollout timing (first evaluation per org), and stale flags are all
 * visible. Call in place of `!!org?.features?.[flag]`.
 */
export const useOrgFeatureFlag = (flag: string) => {
  const { org } = useOrg()
  const enabled = !!org?.features?.[flag]

  useEffect(() => {
    if (!org?.id) return
    const key = `${org.id}:${flag}`
    if (evaluated.has(key)) return
    evaluated.add(key)
    captureFeatureFlagEvaluation({ flag, enabled })
  }, [org?.id, flag, enabled])

  return enabled
}
