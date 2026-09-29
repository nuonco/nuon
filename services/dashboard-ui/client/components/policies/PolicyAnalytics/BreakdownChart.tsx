import {
  ResponsiveContainer,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
} from 'recharts'
import { Text } from '@/components/common/Text'
import type { TPolicyAnalyticsBreakdown } from '@/types'
import { ChartLegend } from './ChartLegend'
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

interface IBreakdownChart {
  breakdown: TPolicyAnalyticsBreakdown | undefined
  title: string
  formatLabel?: (key: string) => string
  xMax?: number
}

function truncateLabel(label: string) {
  if (label.length <= 20) return label
  return `${label.slice(0, 9)}…${label.slice(-8)}`
}

const OWNER_TYPE_LABELS: Record<string, string> = {
  install_deploys: 'Deploy',
  component_builds: 'Build',
  install_sandbox_runs: 'Sandbox',
}

function defaultFormatLabel(key: string) {
  return OWNER_TYPE_LABELS[key] ?? key
}

const BAR_SIZE = 22
const BAR_CATEGORY_GAP = 18
const Y_AXIS_WIDTH = 168
const Y_AXIS_LABEL_INSET = 12
const CHART_TOP_PADDING = 8
const X_AXIS_AREA = 28

const LEGEND_ITEMS = [
  { color: CHART_SERIES_FAIL, label: 'Denied' },
  { color: CHART_SERIES_WARN, label: 'Warnings' },
  { color: CHART_SERIES_PASS, label: 'Passed' },
]

interface IYAxisTickProps {
  x?: number | string
  y?: number | string
  payload?: { value?: string | number }
}

const renderYAxisTick = ({ x = 0, y = 0, payload }: IYAxisTickProps) => {
  const numericX = typeof x === 'number' ? x : Number(x) || 0
  const labelX = numericX - Y_AXIS_WIDTH + Y_AXIS_LABEL_INSET
  const value = String(payload?.value ?? '')
  return (
    <text
      x={labelX}
      y={y}
      dy={4}
      textAnchor="start"
      style={chartAxisTickStyle as React.CSSProperties}
    >
      <title>{value}</title>
      {truncateLabel(value)}
    </text>
  )
}

export const BreakdownChart = ({
  breakdown,
  title,
  formatLabel = defaultFormatLabel,
  xMax,
}: IBreakdownChart) => {
  const entries = breakdown?.entries ?? []

  const data = entries.map((e) => ({
    name: formatLabel(e.key ?? ''),
    denies: e.denies ?? 0,
    warns: e.warns ?? 0,
    passes: e.passes ?? 0,
  }))

  const chartHeight = data.length
    ? data.length * (BAR_SIZE + BAR_CATEGORY_GAP) +
      CHART_TOP_PADDING +
      X_AXIS_AREA
    : 120

  return (
    <div className="flex flex-col">
      <div className="flex items-center justify-between gap-4 flex-wrap p-4 border-b border-cool-grey-200 dark:border-dark-grey-600">
        <Text weight="strong" variant="body">
          {title}
        </Text>
        <ChartLegend items={LEGEND_ITEMS} />
      </div>
      <div className="p-4">
        {data.length ? (
          <ResponsiveContainer width="100%" height={chartHeight}>
            <BarChart
              data={data}
              layout="vertical"
              margin={{ top: CHART_TOP_PADDING, right: 8, bottom: 0, left: 0 }}
              barCategoryGap={BAR_CATEGORY_GAP}
            >
              <CartesianGrid
                strokeDasharray="3 3"
                stroke={chartGridStroke}
                opacity={0.5}
                horizontal={false}
              />
              <XAxis
                type="number"
                domain={xMax ? [0, xMax] : undefined}
                tick={chartAxisTickStyle}
                tickLine={false}
                axisLine={false}
                allowDecimals={false}
              />
              <YAxis
                type="category"
                dataKey="name"
                tick={renderYAxisTick}
                tickLine={false}
                axisLine={false}
                width={Y_AXIS_WIDTH}
                interval={0}
              />
              <Tooltip
                contentStyle={chartTooltipContentStyle}
                labelStyle={chartTooltipLabelStyle}
                itemStyle={chartTooltipItemStyle}
                cursor={{ fill: 'var(--foreground)', fillOpacity: 0.06 }}
                allowEscapeViewBox={{ x: true, y: true }}
                wrapperStyle={{ zIndex: 20 }}
              />
              <Bar
                dataKey="denies"
                name="Denied"
                stackId="1"
                fill={CHART_SERIES_FAIL}
                fillOpacity={0.85}
                barSize={BAR_SIZE}
              />
              <Bar
                dataKey="warns"
                name="Warnings"
                stackId="1"
                fill={CHART_SERIES_WARN}
                fillOpacity={0.85}
                barSize={BAR_SIZE}
              />
              <Bar
                dataKey="passes"
                name="Passed"
                stackId="1"
                fill={CHART_SERIES_PASS}
                fillOpacity={0.85}
                barSize={BAR_SIZE}
              />
            </BarChart>
          </ResponsiveContainer>
        ) : (
          <div
            className="flex items-center justify-center"
            style={{ height: chartHeight }}
          >
            <Text variant="subtext" theme="neutral">
              No data
            </Text>
          </div>
        )}
      </div>
    </div>
  )
}
