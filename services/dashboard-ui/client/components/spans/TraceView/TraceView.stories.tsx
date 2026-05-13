import { useState } from 'react'
import type { TSpan } from '@/types'
import { TraceView } from './TraceView'

export default {
  title: 'Spans/TraceView',
}

const mockSpans: TSpan[] = [
  {
    span_id: 'root',
    name: 'job.deploy',
    start_time: '2026-01-01T00:00:00.000Z',
    end_time: '2026-01-01T00:01:00.000Z',
    duration_ns: 60_000_000_000,
    status_code: 'Ok',
    attributes: { 'nuon.tool': 'runner', 'nuon.job.type': 'deploy' },
  },
  {
    span_id: 'fetch-step',
    parent_span_id: 'root',
    name: 'step.fetching',
    start_time: '2026-01-01T00:00:01.000Z',
    end_time: '2026-01-01T00:00:02.000Z',
    duration_ns: 1_000_000_000,
    status_code: 'Ok',
    attributes: { 'nuon.tool': 'runner' },
  },
  {
    span_id: 'git-clone',
    parent_span_id: 'fetch-step',
    name: 'git.clone',
    start_time: '2026-01-01T00:00:01.100Z',
    end_time: '2026-01-01T00:00:01.900Z',
    duration_ns: 800_000_000,
    status_code: 'Ok',
    attributes: { 'nuon.tool': 'git', 'nuon.op': 'clone' },
  },
  {
    span_id: 'plan-step',
    parent_span_id: 'root',
    name: 'step.planning',
    start_time: '2026-01-01T00:00:02.000Z',
    end_time: '2026-01-01T00:00:30.000Z',
    duration_ns: 28_000_000_000,
    status_code: 'Ok',
    attributes: { 'nuon.tool': 'runner' },
  },
  {
    span_id: 'tf-plan',
    parent_span_id: 'plan-step',
    name: 'terraform.plan',
    start_time: '2026-01-01T00:00:10.000Z',
    end_time: '2026-01-01T00:00:30.000Z',
    duration_ns: 20_000_000_000,
    status_code: 'Error',
    attributes: { 'nuon.tool': 'terraform', 'nuon.op': 'plan' },
  },
]

const FakeLogs = ({ spanId }: { spanId?: string }) => (
  <div className="p-4 text-sm">
    {spanId
      ? `(mock logs filtered by span_id=${spanId})`
      : '(mock logs — no span selected)'}
  </div>
)

export const Default = () => {
  const [selected, setSelected] = useState<string | undefined>()
  return (
    <div style={{ height: 600 }}>
      <TraceView
        spans={mockSpans}
        selectedSpanId={selected}
        onSelectSpan={setSelected}
        rightPane={<FakeLogs spanId={selected} />}
      />
    </div>
  )
}

export const Empty = () => (
  <div style={{ height: 600 }}>
    <TraceView
      spans={[]}
      onSelectSpan={() => undefined}
      rightPane={<FakeLogs />}
    />
  </div>
)

export const Loading = () => (
  <div style={{ height: 600 }}>
    <TraceView
      spans={[]}
      isLoading
      onSelectSpan={() => undefined}
      rightPane={<FakeLogs />}
    />
  </div>
)
