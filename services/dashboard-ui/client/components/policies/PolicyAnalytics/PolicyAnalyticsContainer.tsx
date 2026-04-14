import { useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { DateTime } from 'luxon'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getPolicyAnalyticsSummary, getPolicyAnalyticsTimeseries } from '@/lib'
import { PolicyAnalytics } from './PolicyAnalytics'

const RANGES: Record<string, number> = {
  '7d': 7,
  '30d': 30,
  '90d': 90,
  '1y': 365,
}

export const PolicyAnalyticsContainer = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const [selectedRange, setSelectedRange] = useState('30d')

  const { start, end } = useMemo(() => {
    const now = DateTime.now().toUTC().startOf('minute')
    return {
      end: now.toISO()!,
      start: now.minus({ days: RANGES[selectedRange] }).toISO()!,
    }
  }, [selectedRange])

  const { data: summary, isLoading: isLoadingSummary } = useQuery({
    queryKey: ['policy-analytics-summary', org?.id, app?.id, selectedRange],
    queryFn: () =>
      getPolicyAnalyticsSummary({
        orgId: org.id,
        appId: app.id,
        start,
        end,
      }),
    enabled: !!org?.id && !!app?.id,
  })

  const { data: timeseries, isLoading: isLoadingTimeseries } = useQuery({
    queryKey: ['policy-analytics-timeseries', org?.id, app?.id, selectedRange],
    queryFn: () =>
      getPolicyAnalyticsTimeseries({
        orgId: org.id,
        appId: app.id,
        start,
        end,
      }),
    enabled: !!org?.id && !!app?.id,
  })

  return (
    <PolicyAnalytics
      summary={summary}
      timeseries={timeseries}
      isLoading={isLoadingSummary || isLoadingTimeseries}
      selectedRange={selectedRange}
      onRangeChange={setSelectedRange}
    />
  )
}
