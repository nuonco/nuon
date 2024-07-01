import { afterAll, expect, test, vi } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { getGetInstall200Response } from '@test/mock-api-handlers'
import { useInstallContext, InstallProvider } from './install-context'

const mockInstall = getGetInstall200Response()
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

test('useInstallContext should throw error when used outside of InstallProvider', () => {
  try {
    renderHook(() => useInstallContext())
  } catch (error) {
    expect(error).toMatchInlineSnapshot(
      `[Error: useInstallContext() may only be used in the context of a <InstallProvider> component.]`
    )
  }
})

test('install context should render with init state', () => {
  const { result } = renderHook(() => useInstallContext(), {
    wrapper: ({ children }) => (
      <InstallProvider initInstall={mockInstall}>{children}</InstallProvider>
    ),
  })
  expect(result.current.install).toHaveProperty('id', mockInstall.id)
  expect(result.current.install).toHaveProperty('name', mockInstall.name)
  expect(result.current.isFetching).toBeFalsy()
})

test.skip(
  'install context should refetch it state from api if provider has polling enabled',
  async () => {
    const { result } = renderHook(() => useInstallContext(), {
      wrapper: ({ children }) => (
        <InstallProvider initInstall={mockInstall} shouldPoll>
          {children}
        </InstallProvider>
      ),
    })
    expect(result.current.install).toHaveProperty('id', mockInstall.id)
    expect(result.current.install).toHaveProperty('name', mockInstall.name)
    expect(result.current.isFetching).toBeFalsy()

    // starting install refetch
    await waitFor(
      () => {
        return expect(result.current.isFetching).toBeTruthy()
      },
      {
        timeout: POLL_DURATION + 1000,
      }
    )

    // finish install refetch
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
