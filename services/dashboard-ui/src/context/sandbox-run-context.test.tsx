import { afterAll, expect, test, vi } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { getGetInstallSandboxRuns200Response } from '@test/mock-api-handlers'
import { useSandboxRunContext, SandboxRunProvider } from './sandbox-run-context'

const mockSandboxRun = getGetInstallSandboxRuns200Response()[0]
vi.mock('../utils', async (og) => {
  const mod = await og<typeof import('../utils')>()
  return {
    ...mod,
    POLL_DURATION: 2000,
  }
})

import { POLL_DURATION } from '../utils'

afterAll(() => {
  vi.resetAllMocks()
})

test('useSandboxRunContext should throw error when used outside of SandboxRunProvider', () => {
  try {
    renderHook(() => useSandboxRunContext())
  } catch (error) {
    expect(error).toMatchInlineSnapshot(
      `[Error: useSandboxRunContext() may only be used in the context of a <SandboxRunProvider> component.]`
    )
  }
})

test('sandbox-run context should render with init state', () => {
  const { result } = renderHook(() => useSandboxRunContext(), {
    wrapper: ({ children }) => (
      <SandboxRunProvider initRun={mockSandboxRun}>
        {children}
      </SandboxRunProvider>
    ),
  })
  expect(result.current.run).toHaveProperty('id', mockSandboxRun.id)
  expect(result.current.run).toHaveProperty(
    'install_id',
    mockSandboxRun.install_id
  )
  expect(result.current.isFetching).toBeFalsy()
})

test.skip(
  'sandbox-run context should refetch it state from api if provider has polling enabled',
  async () => {
    const { result } = renderHook(() => useSandboxRunContext(), {
      wrapper: ({ children }) => (
        <SandboxRunProvider
          initRun={{ ...mockSandboxRun, org_id: 'test' }}
          shouldPoll
        >
          {children}
        </SandboxRunProvider>
      ),
    })
    expect(result.current.run).toHaveProperty('id', mockSandboxRun.id)
    expect(result.current.run).toHaveProperty(
      'install_id',
      mockSandboxRun.install_id
    )
    expect(result.current.isFetching).toBeFalsy()

    // starting sandbox-run refetch
    await waitFor(
      () => {
        return expect(result.current.isFetching).toBeTruthy()
      },
      {
        timeout: POLL_DURATION + 1000,
      }
    )

    // finish sandbox-run refetch
    await waitFor(
      () => {
        return expect(result.current.isFetching).toBeFalsy()
      },
      {
        timeout: POLL_DURATION + 1001,
      }
    )
  },
  POLL_DURATION * 1050
)
