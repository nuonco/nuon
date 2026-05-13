import { useMemo } from 'react'
import {
  ResponsiveContainer,
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
} from 'recharts'
import { DateTime } from 'luxon'
import { Text } from '@/components/common/Text'
import type { TPolicyAnalyticsTimeseries } from '@/types'
import {
  CHART_SERIES_FAIL,
  CHART_SERIES_PASS,
  CHART_SERIES_WARN,
  chartAxisTickStyle,
  chartGridStroke,
  chartTooltipContentStyle,
  chartTooltipItemStyle,
  chartTooltipLabelStyle,
} from './chart-theme'

interface IPolicyAnalyticsChart {
  timeseries: TPolicyAnalyticsTimeseries | undefined
}

function formatTick(value: string, interval: string) {
  const dt = DateTime.fromISO(value)
  if (!dt.isValid) return value
  if (interval === '15m' || interval === '30m') {
    return dt.toFormat('HH:mm')
  }
  if (interval === 'hour' || interval === '6h') {
    return dt.toFormat('MMM d HH:mm')
  }
  return dt.toFormat('MMM d')
}

function formatTooltipLabel(value: string) {
  const dt = DateTime.fromISO(value)
  if (!dt.isValid) return value
  return dt.toFormat('MMM d, yyyy HH:mm')
}

export const PolicyAnalyticsChart = ({ timeseries }: IPolicyAnalyticsChart) => {
  const data = useMemo(() => {
    if (!timeseries?.buckets?.length) return []
    return timeseries.buckets.map((b) => ({
      time: b.time,
      passes: b.passes,
      warns: b.warns,
      denies: b.denies,
    }))
  }, [timeseries])

  if (!data.length) {
    return (
      <div className="flex items-center justify-center h-[300px]">
        <Text variant="subtext" theme="neutral">
          No evaluation data for this time range
        </Text>
      </div>
    )
  }

  const interval = timeseries?.interval ?? 'day'

  return (
    <ResponsiveContainer width="100%" height={300}>
      <AreaChart data={data} margin={{ top: 8, right: 8, bottom: 0, left: 0 }}>
        <CartesianGrid
          strokeDasharray="3 3"
          stroke={chartGridStroke}
          opacity={0.5}
        />
        <XAxis
          dataKey="time"
          tickFormatter={(v) => formatTick(v, interval)}
          tick={chartAxisTickStyle}
          tickLine={false}
          axisLine={false}
        />
        <YAxis
          tick={chartAxisTickStyle}
          tickLine={false}
          axisLine={false}
          allowDecimals={false}
        />
        <Tooltip
          labelFormatter={formatTooltipLabel}
          contentStyle={chartTooltipContentStyle}
          labelStyle={chartTooltipLabelStyle}
          itemStyle={chartTooltipItemStyle}
          cursor={{ stroke: 'var(--foreground)', strokeOpacity: 0.2 }}
          allowEscapeViewBox={{ x: true, y: true }}
          wrapperStyle={{ zIndex: 20 }}
        />
        <Area
          type="monotone"
          dataKey="passes"
          name="Passed"
          stackId="1"
          stroke={CHART_SERIES_PASS}
          fill={CHART_SERIES_PASS}
          fillOpacity={0.3}
          strokeWidth={2}
        />
        <Area
          type="monotone"
          dataKey="warns"
          name="Warnings"
          stackId="1"
          stroke={CHART_SERIES_WARN}
          fill={CHART_SERIES_WARN}
          fillOpacity={0.3}
          strokeWidth={2}
        />
        <Area
          type="monotone"
          dataKey="denies"
          name="Denied"
          stackId="1"
          stroke={CHART_SERIES_FAIL}
          fill={CHART_SERIES_FAIL}
          fillOpacity={0.3}
          strokeWidth={2}
        />
      </AreaChart>
    </ResponsiveContainer>
  )
}
