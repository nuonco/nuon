import { useEffect } from 'react'
import { captureFeatureFlagEvaluation } from '@/lib/posthog-analytics'
import { useOrg } from '@/hooks/use-org'

const evaluated = new Set<string>()

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
