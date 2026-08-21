import { useMemo } from 'react'
import {
  ResponsiveContainer,
  ComposedChart,
  Line,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
} from 'recharts'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Text } from '@/components/common/Text'
import {
  chartAxisTickStyle,
  chartGridStroke,
  chartTooltipContentStyle,
  chartTooltipItemStyle,
  chartTooltipLabelStyle,
} from '@/components/policies/PolicyAnalytics/chart-theme'
import { parseXYChart } from './parse-xychart'

// Series colors follow the categorical order used across the dashboard;
// most runbook charts are single-series and only use the first slot.
const SERIES_COLORS = [
  'var(--color-blue-500)',
  'var(--color-orange-500)',
  'var(--color-green-500)',
  'var(--color-red-500)',
]

interface IMermaidXYChart {
  code: string
}

/**
 * Renders a mermaid `xychart-beta` block as a Recharts chart, falling back to
 * the highlighted source when the block doesn't parse.
 */
export const MermaidXYChart = ({ code }: IMermaidXYChart) => {
  const chart = useMemo(() => parseXYChart(code), [code])

  const data = useMemo(() => {
    if (!chart) return []
    const length = Math.max(...chart.series.map((s) => s.values.length))
    return Array.from({ length }, (_, i) => {
      const row: Record<string, number | string> = {
        x: chart.xLabels?.[i] ?? i + 1,
      }
      chart.series.forEach((s, si) => {
        row[`s${si}`] = s.values[i]
      })
      return row
    })
  }, [chart])

  if (!chart || !data.length) return <CodeBlock language="mermaid">{code}</CodeBlock>

  return (
    <div className="my-4">
      {chart.title ? (
        <Text variant="subtext" theme="default" className="mb-1">
          {chart.title}
        </Text>
      ) : null}
      <ResponsiveContainer width="100%" height={220}>
        <ComposedChart data={data} margin={{ top: 8, right: 8, bottom: 0, left: 0 }}>
          <CartesianGrid strokeDasharray="3 3" stroke={chartGridStroke} opacity={0.5} />
          <XAxis dataKey="x" tick={chartAxisTickStyle} tickLine={false} axisLine={false} minTickGap={24} />
          <YAxis
            tick={chartAxisTickStyle}
            tickLine={false}
            axisLine={false}
            width={48}
            domain={[
              chart.yMin ?? 'auto',
              chart.yMax ?? 'auto',
            ]}
            label={
              chart.yLabel
                ? { value: chart.yLabel, angle: -90, position: 'insideLeft', style: chartAxisTickStyle }
                : undefined
            }
          />
          <Tooltip
            contentStyle={chartTooltipContentStyle}
            labelStyle={chartTooltipLabelStyle}
            itemStyle={chartTooltipItemStyle}
            cursor={{ stroke: 'var(--foreground)', strokeOpacity: 0.2 }}
            wrapperStyle={{ zIndex: 20 }}
          />
          {chart.series.map((s, si) =>
            s.kind === 'bar' ? (
              <Bar
                key={si}
                dataKey={`s${si}`}
                name={chart.title ?? `series ${si + 1}`}
                fill={SERIES_COLORS[si % SERIES_COLORS.length]}
                fillOpacity={0.6}
                radius={[4, 4, 0, 0]}
              />
            ) : (
              <Line
                key={si}
                type="monotone"
                dataKey={`s${si}`}
                name={chart.title ?? `series ${si + 1}`}
                stroke={SERIES_COLORS[si % SERIES_COLORS.length]}
                strokeWidth={2}
                dot={false}
              />
            ),
          )}
        </ComposedChart>
      </ResponsiveContainer>
    </div>
  )
}
