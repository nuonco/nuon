import { useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { DateTime } from 'luxon'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import {
  getAppInstalls,
  getAppPoliciesConfigs,
  getPolicyAnalyticsBreakdown,
  getPolicyAnalyticsSummary,
  getPolicyAnalyticsTimeseries,
} from '@/lib'
import { PolicyAnalytics } from './PolicyAnalytics'

const RANGES: Record<string, { hours: number }> = {
  '2h': { hours: 2 },
  '24h': { hours: 24 },
  '7d': { hours: 7 * 24 },
  '30d': { hours: 30 * 24 },
  '90d': { hours: 90 * 24 },
  '1y': { hours: 365 * 24 },
}

export const PolicyAnalyticsContainer = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const [selectedRange, setSelectedRange] = useState('30d')

  const { start, end } = useMemo(() => {
    const now = DateTime.now().toUTC().startOf('minute')
    return {
      end: now.toISO()!,
      start: now.minus({ hours: RANGES[selectedRange].hours }).toISO()!,
    }
  }, [selectedRange])

  const enabled = !!org?.id && !!app?.id
  const baseKey = [org?.id, app?.id, selectedRange]

  const { data: summary, isLoading: isLoadingSummary } = useQuery({
    queryKey: ['policy-analytics-summary', ...baseKey],
    queryFn: () =>
      getPolicyAnalyticsSummary({ orgId: org.id, appId: app.id, start, end }),
    enabled,
  })

  const { data: timeseries, isLoading: isLoadingTimeseries } = useQuery({
    queryKey: ['policy-analytics-timeseries', ...baseKey],
    queryFn: () =>
      getPolicyAnalyticsTimeseries({
        orgId: org.id,
        appId: app.id,
        start,
        end,
      }),
    enabled,
  })

  const { data: byPolicy } = useQuery({
    queryKey: ['policy-analytics-breakdown', 'policy_id', ...baseKey],
    queryFn: () =>
      getPolicyAnalyticsBreakdown({
        orgId: org.id,
        appId: app.id,
        dimension: 'policy_id',
        start,
        end,
      }),
    enabled,
  })

  const { data: byInstall } = useQuery({
    queryKey: ['policy-analytics-breakdown', 'install_id', ...baseKey],
    queryFn: () =>
      getPolicyAnalyticsBreakdown({
        orgId: org.id,
        appId: app.id,
        dimension: 'install_id',
        start,
        end,
      }),
    enabled,
  })

  const { data: byOwnerType } = useQuery({
    queryKey: ['policy-analytics-breakdown', 'owner_type', ...baseKey],
    queryFn: () =>
      getPolicyAnalyticsBreakdown({
        orgId: org.id,
        appId: app.id,
        dimension: 'owner_type',
        start,
        end,
      }),
    enabled,
  })

  const { data: installs } = useQuery({
    queryKey: ['app-installs', org?.id, app?.id],
    queryFn: () => getAppInstalls({ orgId: org.id, appId: app.id }),
    enabled,
  })

  const { data: policiesConfigs } = useQuery({
    queryKey: ['app-policies-configs', org?.id, app?.id],
    queryFn: () => getAppPoliciesConfigs({ orgId: org.id, appId: app.id }),
    enabled,
  })

  const installNames = useMemo(() => {
    const map: Record<string, string> = {}
    for (const inst of installs?.data ?? []) {
      if (inst.id && inst.name) map[inst.id] = inst.name
    }
    return map
  }, [installs])

  const policyNames = useMemo(() => {
    const map: Record<string, string> = {}
    for (const cfg of policiesConfigs ?? []) {
      for (const p of cfg.policies ?? []) {
        if (p.id && p.name) map[p.id] = p.name
      }
    }
    return map
  }, [policiesConfigs])

  return (
    <PolicyAnalytics
      summary={summary}
      timeseries={timeseries}
      byPolicy={byPolicy}
      byInstall={byInstall}
      byOwnerType={byOwnerType}
      policyNames={policyNames}
      installNames={installNames}
      isLoading={isLoadingSummary || isLoadingTimeseries}
      selectedRange={selectedRange}
      onRangeChange={setSelectedRange}
    />
  )
}
