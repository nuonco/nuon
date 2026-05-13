import { useMemo, useState } from 'react'
import type { TSpan } from '@/types'
import { collectSpanIds, SpanTree } from './SpanTree'

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
    attributes: {
      'nuon.tool': 'terraform',
      'nuon.op': 'init',
      'runner_job.id': 'jobe37y0720x19xezfgqzifskw',
      'runner_job_execution.id': 'rjeabc123',
      'runner_job_execution_step.name': 'plan',
      'terraform.version': '1.7.5',
    },
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
    attributes: {
      'nuon.tool': 'terraform',
      'nuon.op': 'plan',
      'runner_job.id': 'jobe37y0720x19xezfgqzifskw',
      'runner_job_execution.id': 'rjeabc123',
      'runner_job_execution_step.name': 'plan',
      'terraform.workspace': 'default',
      'terraform.exit_code': '1',
    },
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

const useTreeProps = (spans: TSpan[]) => {
  const allIds = useMemo(() => collectSpanIds(spans), [spans])
  const [collapsed, setCollapsed] = useState<Set<string>>(new Set())
  return {
    collapsed,
    onToggleCollapsed: (id: string) =>
      setCollapsed((prev) => {
        const next = new Set(prev)
        if (next.has(id)) next.delete(id)
        else next.add(id)
        return next
      }),
    onExpandAll: () => setCollapsed(new Set()),
    onCollapseAll: () => setCollapsed(new Set(allIds)),
  }
}

export const Default = () => {
  const [selected, setSelected] = useState<string | undefined>('tf-plan')
  const treeProps = useTreeProps(mockSpans)
  return (
    <div style={{ width: 540 }}>
      <SpanTree
        spans={mockSpans}
        selectedSpanId={selected}
        onSelectSpan={setSelected}
        {...treeProps}
      />
    </div>
  )
}

export const Empty = () => {
  const treeProps = useTreeProps([])
  return (
    <SpanTree spans={[]} onSelectSpan={() => undefined} {...treeProps} />
  )
}

export const SingleRoot = () => {
  const [selected, setSelected] = useState<string | undefined>()
  const single = [mockSpans[0]]
  const treeProps = useTreeProps(single)
  return (
    <SpanTree
      spans={single}
      selectedSpanId={selected}
      onSelectSpan={setSelected}
      {...treeProps}
    />
  )
}
