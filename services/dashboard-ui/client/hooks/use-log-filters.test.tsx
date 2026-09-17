import { expect, test } from 'bun:test'
import { cleanup, render } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import type { TOTELLog } from '@/types'
import { buildServerFilters, useLogFilters } from './use-log-filters'

const toolLog = (id: string, tool: string): TOTELLog =>
  ({
    id,
    timestamp: `2026-01-01T00:00:0${id}.000Z`,
    body: `line ${id}`,
    severity_text: 'Info',
    log_attributes: { 'nuon.tool': tool },
  }) as unknown as TOTELLog

let latest: ReturnType<typeof useLogFilters> | undefined

const HookProbe = ({
  logs,
  streamId,
}: {
  logs: TOTELLog[]
  streamId?: string
}) => {
  latest = useLogFilters(logs, undefined, streamId)
  return null
}

const renderHook = (logs: TOTELLog[], search: string, streamId?: string) => {
  const view = render(
    <MemoryRouter initialEntries={[`/logs${search}`]}>
      <HookProbe logs={logs} streamId={streamId} />
    </MemoryRouter>
  )
  return {
    rerender: (nextLogs: TOTELLog[], nextStreamId = streamId) => {
      view.rerender(
        <MemoryRouter initialEntries={[`/logs${search}`]}>
          <HookProbe logs={nextLogs} streamId={nextStreamId} />
        </MemoryRouter>
      )
    },
  }
}

test('buildServerFilters maps repeated, scoped, trimmed and exact params', () => {
  const sp = new URLSearchParams(
    '?severity=Error&severity=Warn&system_logs=false&tool=helm&q=%20%20boom%20&trace_id=t1'
  )
  expect(buildServerFilters(sp)).toStrictEqual({
    severity_text: ['Error', 'Warn'],
    scope_name: ['oteljob'],
    tool: 'helm',
    q: 'boom',
    trace_id: 't1',
  })
})

test('buildServerFilters applies default severities when none are set', () => {
  const sp = new URLSearchParams('?q=boom')
  expect(buildServerFilters(sp).severity_text).toStrictEqual([
    'Info',
    'Warn',
    'Error',
    'Fatal',
  ])
})

test('useLogFilters serverFilters matches buildServerFilters for the same URL', () => {
  const logs = [toolLog('1', 'terraform')]
  renderHook(logs, '?severity=Error&system_logs=false&tool=helm&q=%20boom%20')
  expect(latest?.serverFilters).toStrictEqual(
    buildServerFilters(new URLSearchParams('?severity=Error&system_logs=false&tool=helm&q=%20boom%20'))
  )
  cleanup()
})

test('tool facets survive a server-side tool filter', () => {
  const logs = [toolLog('1', 'helm'), toolLog('2', 'terraform')]
  renderHook(logs, '?tool=helm', 'ls1')

  expect(latest?.tool).toBe('helm')
  expect(latest?.availableTools.has('helm')).toBe(true)
  expect(latest?.availableTools.has('terraform')).toBe(true)
  cleanup()
})

test('tool facets survive cleared logs on filter reconnect', () => {
  // Pick tool=helm: reconnect clears logs, then only helm rows arrive.
  const { rerender } = renderHook(
    [toolLog('1', 'helm'), toolLog('2', 'terraform')],
    '?tool=helm',
    'ls1'
  )
  expect(latest?.availableTools.has('terraform')).toBe(true)

  // Provider's connect effect calls setLogs([]) — facets must not reset.
  rerender([])
  expect(latest?.availableTools.has('helm')).toBe(true)
  expect(latest?.availableTools.has('terraform')).toBe(true)

  // Refilled with only the selected tool's rows.
  rerender([toolLog('3', 'helm')])
  expect(latest?.availableTools.has('helm')).toBe(true)
  expect(latest?.availableTools.has('terraform')).toBe(true)
  cleanup()
})

test('tool facets reset when the stream changes', () => {
  const logs = [toolLog('1', 'helm')]
  const { rerender } = renderHook(logs, '?tool=oci', 'ls1')
  expect(latest?.availableTools.has('helm')).toBe(true)

  // The first render after a stream change may still carry the old
  // stream's logs; they must not leak into the new stream's facets.
  rerender([toolLog('2', 'oci')], 'ls2')
  expect(latest?.availableTools.size).toBe(0)

  // Accumulation resumes from the new stream's own rows.
  rerender([toolLog('3', 'oci')], 'ls2')
  expect(latest?.availableTools.has('oci')).toBe(true)
  expect(latest?.availableTools.has('helm')).toBe(false)
  cleanup()
})

test('stream change records the new id even when logs are empty', () => {
  const { rerender } = renderHook([toolLog('1', 'helm')], '', 'ls1')
  expect(latest?.availableTools.has('helm')).toBe(true)

  // Stale render: id change and old-stream logs land together.
  rerender([toolLog('2', 'helm')], 'ls2')
  // Provider clears logs before the new stream's first batch.
  rerender([], 'ls2')
  // Closed stream delivered in a single batch: its tools must survive —
  // the reset must not fire again on this real data.
  rerender([toolLog('3', 'oci')], 'ls2')
  expect(latest?.availableTools.size).toBe(1)
  expect(latest?.availableTools.has('oci')).toBe(true)
  expect(latest?.availableTools.has('helm')).toBe(false)
  cleanup()
})

test('first mount with logs already present accumulates', () => {
  renderHook([toolLog('1', 'helm')], '', 'ls1')
  expect(latest?.availableTools.has('helm')).toBe(true)
  cleanup()
})
