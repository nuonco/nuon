import {
  ResponsiveContainer,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
} from 'recharts'
import { Text } from '@/components/common/Text'
import type { TPolicyAnalyticsBreakdown } from '@/types'

interface IBreakdownChart {
  breakdown: TPolicyAnalyticsBreakdown | undefined
  title: string
  formatLabel?: (key: string) => string
}

function truncateId(id: string) {
  if (id.length <= 16) return id
  return `${id.slice(0, 6)}…${id.slice(-6)}`
}

const OWNER_TYPE_LABELS: Record<string, string> = {
  install_deploys: 'Deploy',
  component_builds: 'Build',
  install_sandbox_runs: 'Sandbox',
}

function defaultFormatLabel(key: string) {
  return OWNER_TYPE_LABELS[key] ?? truncateId(key)
}

export const BreakdownChart = ({
  breakdown,
  title,
  formatLabel = defaultFormatLabel,
}: IBreakdownChart) => {
  const entries = breakdown?.entries ?? []

  if (!entries.length) {
    return (
      <div className="flex items-center justify-center h-full min-h-[120px]">
        <Text variant="subtext" theme="neutral">
          No data
        </Text>
      </div>
    )
  }

  const data = entries.map((e) => ({
    name: formatLabel(e.key ?? ''),
    denies: e.denies ?? 0,
    warns: e.warns ?? 0,
    passes: e.passes ?? 0,
  }))

  return (
    <div className="flex flex-col gap-3">
      <Text weight="strong" variant="body">
        {title}
      </Text>
      <ResponsiveContainer width="100%" height={Math.max(120, data.length * 40 + 40)}>
        <BarChart
          data={data}
          layout="vertical"
          margin={{ top: 0, right: 8, bottom: 0, left: 0 }}
        >
          <CartesianGrid strokeDasharray="3 3" opacity={0.3} horizontal={false} />
          <XAxis type="number" tick={{ fontSize: 12 }} tickLine={false} axisLine={false} allowDecimals={false} />
          <YAxis
            type="category"
            dataKey="name"
            tick={{ fontSize: 12 }}
            tickLine={false}
            axisLine={false}
            width={120}
          />
          <Tooltip
            contentStyle={{
              backgroundColor: 'var(--color-surface, #fff)',
              border: '1px solid var(--color-border, #e5e7eb)',
              borderRadius: '8px',
              fontSize: '13px',
            }}
          />
          <Legend wrapperStyle={{ fontSize: '12px' }} />
          <Bar dataKey="denies" name="Denied" stackId="1" fill="#ef4444" fillOpacity={0.85} />
          <Bar dataKey="warns" name="Warnings" stackId="1" fill="#f97316" fillOpacity={0.85} />
          <Bar dataKey="passes" name="Passed" stackId="1" fill="#22c55e" fillOpacity={0.85} />
        </BarChart>
      </ResponsiveContainer>
    </div>
  )
}
