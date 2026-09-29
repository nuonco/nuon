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
  const { rerender } = renderHook(
    [toolLog('1', 'helm'), toolLog('2', 'terraform')],
    '?tool=helm',
    'ls1'
  )
  expect(latest?.availableTools.has('terraform')).toBe(true)

  rerender([])
  expect(latest?.availableTools.has('helm')).toBe(true)
  expect(latest?.availableTools.has('terraform')).toBe(true)

  rerender([toolLog('3', 'helm')])
  expect(latest?.availableTools.has('helm')).toBe(true)
  expect(latest?.availableTools.has('terraform')).toBe(true)
  cleanup()
})

test('tool facets reset when the stream changes', () => {
  const logs = [toolLog('1', 'helm')]
  const { rerender } = renderHook(logs, '?tool=oci', 'ls1')
  expect(latest?.availableTools.has('helm')).toBe(true)

  rerender([toolLog('2', 'oci')], 'ls2')
  expect(latest?.availableTools.size).toBe(0)

  rerender([toolLog('3', 'oci')], 'ls2')
  expect(latest?.availableTools.has('oci')).toBe(true)
  expect(latest?.availableTools.has('helm')).toBe(false)
  cleanup()
})

test('stream change records the new id even when logs are empty', () => {
  const { rerender } = renderHook([toolLog('1', 'helm')], '', 'ls1')
  expect(latest?.availableTools.has('helm')).toBe(true)

  rerender([toolLog('2', 'helm')], 'ls2')
  rerender([], 'ls2')
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
