import { useState } from 'react'
import type { TSpan } from '@/types'
import { SpanTimeline } from './SpanTimeline'

export default {
  title: 'Spans/SpanTimeline',
}

const mockSpans: TSpan[] = [
  {
    span_id: 'root',
    name: 'job.deploy',
    start_time: '2026-01-01T00:00:00.000Z',
    end_time: '2026-01-01T00:01:00.000Z',
    duration_ns: 60_000_000_000,
    status_code: 'Ok',
  },
  {
    span_id: 'init',
    parent_span_id: 'root',
    name: 'step.init',
    start_time: '2026-01-01T00:00:00.000Z',
    end_time: '2026-01-01T00:00:02.000Z',
    duration_ns: 2_000_000_000,
    status_code: 'Ok',
  },
  {
    span_id: 'plan',
    parent_span_id: 'root',
    name: 'step.plan',
    start_time: '2026-01-01T00:00:02.000Z',
    end_time: '2026-01-01T00:00:30.000Z',
    duration_ns: 28_000_000_000,
    status_code: 'Ok',
  },
  {
    span_id: 'tf-plan',
    parent_span_id: 'plan',
    name: 'terraform.plan',
    start_time: '2026-01-01T00:00:10.000Z',
    end_time: '2026-01-01T00:00:30.000Z',
    duration_ns: 20_000_000_000,
    status_code: 'Error',
  },
]

export const Default = () => {
  const [selected, setSelected] = useState<string | undefined>('plan')
  return (
    <div style={{ width: 720 }}>
      <SpanTimeline
        spans={mockSpans}
        selectedSpanId={selected}
        onSelectSpan={setSelected}
      />
    </div>
  )
}

export const Empty = () => (
  <SpanTimeline spans={[]} onSelectSpan={() => undefined} />
)
