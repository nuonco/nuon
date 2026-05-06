import { useState } from 'react'
import type { TSpan } from '@/types'
import { SpanTree } from './SpanTree'

export default {
  title: 'Spans/SpanTree',
}

const mockSpans: TSpan[] = [
  {
    span_id: 'root',
    name: 'job.deploy',
    start_time: '2026-01-01T00:00:00.000Z',
    end_time: '2026-01-01T00:01:00.000Z',
    duration_ns: 60_000_000_000,
    status_code: 'Ok',
    attributes: { 'nuon.tool': 'runner' },
  },
  {
    span_id: 'step-init',
    parent_span_id: 'root',
    name: 'step.init',
    start_time: '2026-01-01T00:00:00.000Z',
    end_time: '2026-01-01T00:00:02.000Z',
    duration_ns: 2_000_000_000,
    status_code: 'Ok',
    attributes: { 'nuon.tool': 'runner' },
  },
  {
    span_id: 'step-plan',
    parent_span_id: 'root',
    name: 'step.plan',
    start_time: '2026-01-01T00:00:02.000Z',
    end_time: '2026-01-01T00:00:30.000Z',
    duration_ns: 28_000_000_000,
    status_code: 'Ok',
    attributes: { 'nuon.tool': 'runner' },
  },
  {
    span_id: 'tf-init',
    parent_span_id: 'step-plan',
    name: 'terraform.init',
    start_time: '2026-01-01T00:00:02.000Z',
    end_time: '2026-01-01T00:00:10.000Z',
    duration_ns: 8_000_000_000,
    status_code: 'Ok',
    attributes: { 'nuon.tool': 'terraform', 'nuon.op': 'init' },
  },
  {
    span_id: 'tf-plan',
    parent_span_id: 'step-plan',
    name: 'terraform.plan',
    start_time: '2026-01-01T00:00:10.000Z',
    end_time: '2026-01-01T00:00:30.000Z',
    duration_ns: 20_000_000_000,
    status_code: 'Error',
    status_message: 'plan failed: invalid resource',
    attributes: { 'nuon.tool': 'terraform', 'nuon.op': 'plan' },
  },
  {
    span_id: 'step-cleanup',
    parent_span_id: 'root',
    name: 'step.cleanup',
    start_time: '2026-01-01T00:00:30.000Z',
    end_time: '2026-01-01T00:00:30.000Z',
    duration_ns: 0,
    status_code: 'Unset',
    attributes: { 'nuon.tool': 'runner' },
  },
]

export const Default = () => {
  const [selected, setSelected] = useState<string | undefined>('tf-plan')
  return (
    <div style={{ width: 540 }}>
      <SpanTree
        spans={mockSpans}
        selectedSpanId={selected}
        onSelectSpan={setSelected}
      />
    </div>
  )
}

export const Empty = () => (
  <SpanTree spans={[]} onSelectSpan={() => undefined} />
)

export const SingleRoot = () => {
  const [selected, setSelected] = useState<string | undefined>()
  return (
    <SpanTree
      spans={[mockSpans[0]]}
      selectedSpanId={selected}
      onSelectSpan={setSelected}
    />
  )
}
