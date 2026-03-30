import { useEffect, useRef } from 'react'
import { useQuery } from '@tanstack/react-query'
import { getCurrentOnboarding } from '@/lib'
import type { TOnboarding } from '@/types'

/**
 * Polls getCurrentOnboarding while step_status is "processing".
 * When the step resolves (idle or error), calls onResolved with the updated onboarding.
 */
export function useOnboardingPoll({
  enabled,
  onResolved,
}: {
  enabled: boolean
  onResolved: (ob: TOnboarding) => void
}) {
  const onResolvedRef = useRef(onResolved)
  onResolvedRef.current = onResolved

  const { data } = useQuery({
    queryKey: ['onboarding-poll'],
    queryFn: getCurrentOnboarding,
    enabled,
    refetchInterval: enabled ? 2000 : false,
  })

  useEffect(() => {
    if (!data || !enabled) return
    if (data.step_status !== 'processing') {
      onResolvedRef.current(data)
    }
  }, [data, enabled])
}
