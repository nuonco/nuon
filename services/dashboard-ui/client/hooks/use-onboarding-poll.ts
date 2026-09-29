import { useEffect, useRef } from 'react'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { getCurrentOnboarding } from '@/lib'
import type { TOnboarding } from '@/types'

export function useOnboardingPoll({
  enabled,
  onResolved,
}: {
  enabled: boolean
  onResolved: (ob: TOnboarding) => void
}) {
  const onResolvedRef = useRef(onResolved)
  onResolvedRef.current = onResolved
  const enabledAtRef = useRef(0)

  useEffect(() => {
    if (enabled) {
      enabledAtRef.current = Date.now()
    }
  }, [enabled])

  const { data, dataUpdatedAt } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['onboarding-poll'],
    queryFn: getCurrentOnboarding,
    enabled,
    refetchInterval: enabled ? 2000 : false,
  })

  useEffect(() => {
    if (!data || !enabled) return
    if (dataUpdatedAt < enabledAtRef.current) return
    if (data.status_v2?.status !== 'in-progress') {
      onResolvedRef.current(data)
    }
  }, [data, dataUpdatedAt, enabled])
}
