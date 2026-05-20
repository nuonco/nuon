import { useMemo } from 'react'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import type {
  TPolicyAnalyticsBreakdown,
  TPolicyAnalyticsSummary,
  TPolicyAnalyticsTimeseries,
} from '@/types'
import { BreakdownChart } from './BreakdownChart'
import {
  CHART_SERIES_FAIL,
  CHART_SERIES_PASS,
  CHART_SERIES_WARN,
} from './chart-theme'
import { ChartLegend } from './ChartLegend'
import { PolicyAnalyticsChart } from './PolicyAnalyticsChart'

const TIMESERIES_LEGEND = [
  { color: CHART_SERIES_FAIL, label: 'Denied' },
  { color: CHART_SERIES_WARN, label: 'Warnings' },
  { color: CHART_SERIES_PASS, label: 'Passed' },
]

/**
 * Compute the largest stacked-bar total across every breakdown so all three
 * BreakdownCharts share a single X-axis scale and bars are visually
 * comparable from one chart to the next.
 */
function computeSharedXMax(
  ...breakdowns: (TPolicyAnalyticsBreakdown | undefined)[]
): number {
  let max = 0
  for (const b of breakdowns) {
    for (const e of b?.entries ?? []) {
      const total = (e.denies ?? 0) + (e.warns ?? 0) + (e.passes ?? 0)
      if (total > max) max = total
    }
  }
  return Math.max(max, 1)
}

const RANGE_OPTIONS = ['2h', '24h', '7d', '30d', '90d', '1y'] as const

function formatInterval(interval: string) {
  switch (interval) {
    case '15m':
      return 'Every 15 min'
    case '30m':
      return 'Every 30 min'
    case 'hour':
      return 'Hourly'
    case '6h':
      return 'Every 6 hours'
    case 'day':
      return 'Daily'
    case 'week':
      return 'Weekly'
    case 'month':
      return 'Monthly'
    default:
      return interval
  }
}

function computePassRate(summary: TPolicyAnalyticsSummary | undefined): string {
  const total = summary?.total_evaluations ?? 0
  if (total === 0) return '—'
  const passes = summary?.total_passes ?? 0
  return `${Math.round((passes / total) * 100)}%`
}

export interface IPolicyAnalytics {
  summary: TPolicyAnalyticsSummary | undefined
  timeseries: TPolicyAnalyticsTimeseries | undefined
  byPolicy: TPolicyAnalyticsBreakdown | undefined
  byInstall: TPolicyAnalyticsBreakdown | undefined
  byOwnerType: TPolicyAnalyticsBreakdown | undefined
  policyNames: Record<string, string>
  installNames: Record<string, string>
  isLoading: boolean
  selectedRange: string
  onRangeChange: (range: string) => void
}

export const PolicyAnalytics = ({
  summary,
  timeseries,
  byPolicy,
  byInstall,
  byOwnerType,
  policyNames,
  installNames,
  isLoading,
  selectedRange,
  onRangeChange,
}: IPolicyAnalytics) => {
  const breakdownXMax = useMemo(
    () => computeSharedXMax(byPolicy, byInstall, byOwnerType),
    [byPolicy, byInstall, byOwnerType]
  )

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center gap-2">
        {RANGE_OPTIONS.map((range) => (
          <Button
            key={range}
            variant={selectedRange === range ? 'primary' : 'secondary'}
            onClick={() => onRangeChange(range)}
            size="sm"
          >
            {range}
          </Button>
        ))}
      </div>

      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        {isLoading ? (
          <>
            <SummaryCardSkeleton />
            <SummaryCardSkeleton />
            <SummaryCardSkeleton />
            <SummaryCardSkeleton />
          </>
        ) : (
          <>
            <Card>
              <div className="flex flex-col gap-1">
                <Text variant="h2" weight="strong">
                  {summary?.total_evaluations ?? 0}
                </Text>
                <Text variant="subtext" theme="neutral">
                  Total evaluations
                </Text>
              </div>
            </Card>
            <Card>
              <div className="flex flex-col gap-1">
                <Text
                  variant="h2"
                  weight="strong"
                  className="text-red-600 dark:text-red-400"
                >
                  {summary?.total_denies ?? 0}
                </Text>
                <Text variant="subtext" theme="neutral">
                  Denied
                </Text>
              </div>
            </Card>
            <Card>
              <div className="flex flex-col gap-1">
                <Text
                  variant="h2"
                  weight="strong"
                  className="text-orange-600 dark:text-orange-400"
                >
                  {summary?.total_warns ?? 0}
                </Text>
                <Text variant="subtext" theme="neutral">
                  Warnings
                </Text>
              </div>
            </Card>
            <Card>
              <div className="flex flex-col gap-1">
                <Text
                  variant="h2"
                  weight="strong"
                  className="text-green-600 dark:text-green-400"
                >
                  {computePassRate(summary)}
                </Text>
                <Text variant="subtext" theme="neutral">
                  Pass rate
                </Text>
              </div>
            </Card>
          </>
        )}
      </div>

      <Card className="!p-0">
        <div className="flex items-center justify-between gap-4 flex-wrap p-4 border-b border-cool-grey-200 dark:border-dark-grey-600">
          <Text weight="strong" variant="body">
            Evaluations over time
          </Text>
          <div className="flex items-center gap-4 flex-wrap">
            <ChartLegend items={TIMESERIES_LEGEND} />
            {timeseries?.interval && (
              <Text variant="subtext" theme="neutral">
                {formatInterval(timeseries.interval)}
              </Text>
            )}
          </div>
        </div>
        <div className="p-4">
          {isLoading ? (
            <Skeleton height="300px" width="100%" />
          ) : (
            <PolicyAnalyticsChart timeseries={timeseries} />
          )}
        </div>
      </Card>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card className="!p-0">
          <BreakdownChart
            breakdown={byPolicy}
            title="Violations by policy"
            formatLabel={(key) => policyNames[key] ?? key}
            xMax={breakdownXMax}
          />
        </Card>
        <Card className="!p-0">
          <BreakdownChart
            breakdown={byInstall}
            title="Violations by install"
            formatLabel={(key) => installNames[key] ?? key}
            xMax={breakdownXMax}
          />
        </Card>
      </div>

      <Card className="!p-0">
        <BreakdownChart
          breakdown={byOwnerType}
          title="Evaluations by stage"
          xMax={breakdownXMax}
        />
      </Card>
    </div>
  )
}

const SummaryCardSkeleton = () => (
  <Card>
    <div className="flex flex-col gap-2">
      <Skeleton height="32px" width="80px" />
      <Skeleton height="16px" width="120px" />
    </div>
  </Card>
)
