import { DateTime } from 'luxon'
import { PolicyAnalytics } from './PolicyAnalytics'
import type { TPolicyAnalyticsSummary, TPolicyAnalyticsTimeseries } from '@/types'

export default { title: 'Policies/PolicyAnalytics' }

const now = DateTime.now().toUTC()

const mockSummary: TPolicyAnalyticsSummary = {
  total_evaluations: 1247,
  total_denies: 23,
  total_warns: 89,
  total_passes: 1135,
  unique_reports: 312,
  unique_policies: 8,
  start: now.minus({ days: 30 }).toISO()!,
  end: now.toISO()!,
}

const mockTimeseries: TPolicyAnalyticsTimeseries = {
  interval: 'day',
  group_by: [],
  start: now.minus({ days: 30 }).toISO()!,
  end: now.toISO()!,
  buckets: Array.from({ length: 15 }, (_, i) => ({
    time: now.minus({ days: 30 - i * 2 }).toISO()!,
    evaluations: 30 + Math.floor(Math.random() * 60),
    passes: 25 + Math.floor(Math.random() * 50),
    warns: Math.floor(Math.random() * 8),
    denies: Math.floor(Math.random() * 3),
  })),
}

export const Default = () => (
  <PolicyAnalytics
    summary={mockSummary}
    timeseries={mockTimeseries}
    isLoading={false}
    selectedRange="30d"
    onRangeChange={() => {}}
  />
)

export const Loading = () => (
  <PolicyAnalytics
    summary={undefined}
    timeseries={undefined}
    isLoading={true}
    selectedRange="30d"
    onRangeChange={() => {}}
  />
)

export const Empty = () => (
  <PolicyAnalytics
    summary={{
      total_evaluations: 0,
      total_denies: 0,
      total_warns: 0,
      total_passes: 0,
      unique_reports: 0,
      unique_policies: 0,
      start: '',
      end: '',
    }}
    timeseries={{
      interval: 'day',
      group_by: [],
      start: '',
      end: '',
      buckets: [],
    }}
    isLoading={false}
    selectedRange="30d"
    onRangeChange={() => {}}
  />
)

export const HighViolations = () => (
  <PolicyAnalytics
    summary={{
      ...mockSummary,
      total_denies: 156,
      total_warns: 234,
      total_passes: 857,
    }}
    timeseries={{
      ...mockTimeseries,
      buckets: mockTimeseries.buckets.map((b) => ({
        ...b,
        denies: 5 + Math.floor(Math.random() * 10),
        warns: 8 + Math.floor(Math.random() * 15),
      })),
    }}
    isLoading={false}
    selectedRange="7d"
    onRangeChange={() => {}}
  />
)
