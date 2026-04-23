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
        <CartesianGrid strokeDasharray="3 3" opacity={0.3} />
        <XAxis
          dataKey="time"
          tickFormatter={(v) => formatTick(v, interval)}
          tick={{ fontSize: 12 }}
          tickLine={false}
          axisLine={false}
        />
        <YAxis
          tick={{ fontSize: 12 }}
          tickLine={false}
          axisLine={false}
          allowDecimals={false}
        />
        <Tooltip
          labelFormatter={formatTooltipLabel}
          contentStyle={{
            backgroundColor: 'var(--color-surface, #fff)',
            border: '1px solid var(--color-border, #e5e7eb)',
            borderRadius: '8px',
            fontSize: '13px',
          }}
        />
        <Area
          type="monotone"
          dataKey="passes"
          name="Passed"
          stackId="1"
          stroke="#22c55e"
          fill="#22c55e"
          fillOpacity={0.3}
          strokeWidth={2}
        />
        <Area
          type="monotone"
          dataKey="warns"
          name="Warnings"
          stackId="1"
          stroke="#f97316"
          fill="#f97316"
          fillOpacity={0.3}
          strokeWidth={2}
        />
        <Area
          type="monotone"
          dataKey="denies"
          name="Denied"
          stackId="1"
          stroke="#ef4444"
          fill="#ef4444"
          fillOpacity={0.3}
          strokeWidth={2}
        />
      </AreaChart>
    </ResponsiveContainer>
  )
}
