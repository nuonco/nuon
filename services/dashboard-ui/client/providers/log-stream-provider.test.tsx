import { afterEach, beforeEach, expect, mock, test } from 'bun:test'
import {
  act,
  cleanup,
  fireEvent,
  render,
  waitFor,
} from '@testing-library/react'
import { useContext } from 'react'
import { MemoryRouter, useSearchParams } from 'react-router'
import * as sseLogs from '@/components/log-stream/SSELogs'
import { OrgContext } from './org-provider'

const getLogStreamLogs = mock(async (_args: unknown): Promise<unknown[]> => [])

mock.module('@/lib/ctl-api/log-streams/get-log-stream-logs', () => ({
  getLogStreamLogs,
}))
mock.module('@/components/log-stream/SSELogs', () => ({
  ...sseLogs,
  LogsPageSkeleton: () => null,
}))

class FakeEventSource {
  static instances: FakeEventSource[] = []

  listeners: Record<string, ((event: unknown) => void)[]> = {}
  onmessage: ((event: unknown) => void) | null = null
  onerror: (() => void) | null = null
  onopen: (() => void) | null = null

  constructor(public url: string) {
    FakeEventSource.instances.push(this)
  }

  addEventListener(type: string, cb: (event: unknown) => void) {
    this.listeners[type] = [...(this.listeners[type] ?? []), cb]
  }

  close() {}

  emitStatus(...data: string[]) {
    act(() => {
      for (const d of data) {
        for (const cb of this.listeners.status ?? []) cb({ data: d })
      }
    })
  }

  emitLogs(logs: unknown[]) {
    act(() => {
      this.onmessage?.({ data: JSON.stringify(logs) })
    })
  }
}

// @ts-expect-error - minimal stand-in for the browser EventSource
globalThis.EventSource = FakeEventSource

const { LogStreamProvider, LogStreamContext } = await import(
  './log-stream-provider'
)

const log = (id: string, runnerJobId?: string) => ({
  id,
  timestamp: `2026-01-01T00:00:0${id}.000Z`,
  body: `line ${id}`,
  runner_job_id: runnerJobId,
})

const Probe = ({
  nextSearch,
}: {
  nextSearch?: string
}) => {
  const ctx = useContext(LogStreamContext)
  const [, setSearchParams] = useSearchParams()
  return (
    <>
      <div data-testid="ids">
        {(ctx?.logs ?? []).map((l) => l.id).join(',')}
      </div>
      {nextSearch !== undefined && (
        <button
          data-testid="set-search"
          onClick={() => setSearchParams(nextSearch)}
        />
      )}
      <button
        data-testid="to-newest-first"
        onClick={() => setSearchParams({ sort: 'desc' })}
      />
    </>
  )
}

const renderProvider = ({
  search = '',
  runnerJobId,
  nextSearch,
}: {
  search?: string
  runnerJobId?: string
  nextSearch?: string
} = {}) => {
  const { getByTestId } = render(
    <MemoryRouter initialEntries={[`/logs${search}`]}>
      <OrgContext.Provider value={{ org: { id: 'org1' }, refresh: () => {} }}>
        <LogStreamProvider logStreamId="ls1" runnerJobId={runnerJobId}>
          <Probe nextSearch={nextSearch} />
        </LogStreamProvider>
      </OrgContext.Provider>
    </MemoryRouter>
  )
  const stream = FakeEventSource.instances.at(-1)
  if (!stream) throw new Error('expected the provider to open a stream')
  return {
    stream,
    ids: () => getByTestId('ids').textContent,
    setSearch: () => fireEvent.click(getByTestId('set-search')),
    switchToNewestFirst: () => fireEvent.click(getByTestId('to-newest-first')),
  }
}

beforeEach(() => {
  FakeEventSource.instances = []
  getLogStreamLogs.mockClear()
  getLogStreamLogs.mockImplementation(async () => [])
})

afterEach(() => {
  cleanup()
})

test('seeds the newest page from the desc read when a newest-first stream is catching up', async () => {
  getLogStreamLogs.mockImplementation(async () => [log('9'), log('8')])

  const { stream, ids } = renderProvider()
  stream.emitLogs([log('1')])
  stream.emitStatus('catching-up')

  await waitFor(() => expect(getLogStreamLogs).toHaveBeenCalledTimes(1))
  expect(getLogStreamLogs.mock.calls[0]?.[0]).toMatchObject({
    logStreamId: 'ls1',
    orgId: 'org1',
    order: 'desc',
  })
  await waitFor(() => expect(ids()).toBe('1,9,8'))
})

test('does not re-add logs the stream already delivered', async () => {
  getLogStreamLogs.mockImplementation(async () => [log('9'), log('1')])

  const { stream, ids } = renderProvider()
  stream.emitLogs([log('1')])
  stream.emitStatus('catching-up')

  await waitFor(() => expect(ids()).toBe('1,9'))
})

test('still seeds when catch-up ends in the same batch it started', async () => {
  getLogStreamLogs.mockImplementation(async () => [log('9')])

  const { stream, ids } = renderProvider()
  stream.emitStatus('catching-up', 'live')

  await waitFor(() => expect(getLogStreamLogs).toHaveBeenCalledTimes(1))
  await waitFor(() => expect(ids()).toBe('9'))
})

test('seeds when the user switches to newest-first mid catch-up', async () => {
  getLogStreamLogs.mockImplementation(async () => [log('9')])

  const { stream, ids, switchToNewestFirst } = renderProvider({
    search: '?sort=asc',
  })
  stream.emitStatus('catching-up')
  expect(getLogStreamLogs).not.toHaveBeenCalled()

  switchToNewestFirst()

  await waitFor(() => expect(getLogStreamLogs).toHaveBeenCalledTimes(1))
  await waitFor(() => expect(ids()).toBe('9'))
})

test('does not seed when the user is sorting oldest-first', async () => {
  const { stream, ids } = renderProvider({ search: '?sort=asc' })
  stream.emitStatus('catching-up')

  await waitFor(() => expect(ids()).toBe(''))
  expect(getLogStreamLogs).not.toHaveBeenCalled()
})

test('does not seed a stream that is already live', async () => {
  const { stream, ids } = renderProvider()
  stream.emitStatus('live')
  stream.emitLogs([log('1')])

  await waitFor(() => expect(ids()).toBe('1'))
  expect(getLogStreamLogs).not.toHaveBeenCalled()
})

test('seeds at most once per stream', async () => {
  const { stream } = renderProvider()
  stream.emitStatus('catching-up')
  await waitFor(() => expect(getLogStreamLogs).toHaveBeenCalledTimes(1))

  stream.emitStatus('live')
  stream.emitStatus('catching-up')

  await waitFor(() => expect(getLogStreamLogs).toHaveBeenCalledTimes(1))
})

test('seeds with a runner job filter for another runner job', async () => {
  getLogStreamLogs.mockImplementation(async () => [
    log('9', 'job1'),
    log('8', 'job2'),
  ])

  const { stream, ids } = renderProvider({ runnerJobId: 'job1' })
  stream.emitStatus('catching-up')

  await waitFor(() => expect(getLogStreamLogs).toHaveBeenCalledTimes(1))
  expect(getLogStreamLogs.mock.calls[0]?.[0]).toMatchObject({
    filters: { runner_job_id: 'job1' },
  })
  await waitFor(() => expect(ids()).toBe('9,8'))
})

test('search keystrokes do not reconnect the stream', () => {
  const { stream, setSearch } = renderProvider({
    search: '?q=a',
    nextSearch: 'q=ab',
  })
  expect(FakeEventSource.instances.length).toBe(1)
  expect(stream.url).not.toContain('q=')

  setSearch()

  expect(FakeEventSource.instances.length).toBe(1)
  expect(stream.url).not.toContain('q=')
})

test('severity change reconnects exactly once with the new filter and resets logs', () => {
  const { stream, ids, setSearch } = renderProvider({
    search: '',
    nextSearch: 'severity=Error',
  })
  stream.emitLogs([log('1')])
  expect(ids()).toBe('1')

  setSearch()

  expect(FakeEventSource.instances.length).toBe(2)
  expect(FakeEventSource.instances[1].url).toContain('severity_text=Error')
  expect(FakeEventSource.instances[1].url).not.toContain('severity_text=Info')
  expect(ids()).toBe('')
})
