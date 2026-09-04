import { afterEach, beforeEach, expect, mock, test } from 'bun:test'
import { act, cleanup, fireEvent, render, waitFor } from '@testing-library/react'
import { useContext } from 'react'
import { MemoryRouter, useSearchParams } from 'react-router'

const getLogStreamLogs = mock(async (_args: unknown): Promise<unknown[]> => [])

mock.module('@/lib/ctl-api/log-streams/get-log-stream-logs', () => ({
  getLogStreamLogs,
}))
mock.module('@/components/log-stream/SSELogs', () => ({
  LogsPageSkeleton: () => null,
}))
mock.module('@/hooks/use-org', () => ({
  useOrg: () => ({ org: { id: 'org1' } }),
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

const Probe = () => {
  const ctx = useContext(LogStreamContext)
  const [, setSearchParams] = useSearchParams()
  return (
    <>
      <div data-testid="ids">
        {(ctx?.logs ?? []).map((l) => l.id).join(',')}
      </div>
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
}: { search?: string; runnerJobId?: string } = {}) => {
  const { getByTestId } = render(
    <MemoryRouter initialEntries={[`/logs${search}`]}>
      <LogStreamProvider logStreamId="ls1" runnerJobId={runnerJobId}>
        <Probe />
      </LogStreamProvider>
    </MemoryRouter>
  )
  const stream = FakeEventSource.instances.at(-1)
  if (!stream) throw new Error('expected the provider to open a stream')
  return {
    stream,
    ids: () => getByTestId('ids').textContent,
    switchToNewestFirst: () =>
      fireEvent.click(getByTestId('to-newest-first')),
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

test('drops seeded logs that belong to another runner job', async () => {
  getLogStreamLogs.mockImplementation(async () => [
    log('9', 'job1'),
    log('8', 'job2'),
  ])

  const { stream, ids } = renderProvider({ runnerJobId: 'job1' })
  stream.emitStatus('catching-up')

  await waitFor(() => expect(ids()).toBe('9'))
})
