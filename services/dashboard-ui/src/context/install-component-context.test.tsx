import { afterAll, expect, test, vi } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { getGetInstallComponents200Response } from '@test/mock-api-handlers'
import {
  useInstallComponentContext,
  InstallComponentProvider,
} from './install-component-context'

const mockInstallComponent = getGetInstallComponents200Response()[0]

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

test('useInstallComponentContext should throw error when used outside of InstallComponentProvider', () => {
  try {
    renderHook(() => useInstallComponentContext())
  } catch (error) {
    expect(error).toMatchInlineSnapshot(
      `[Error: useInstallComponentContext() may only be used in the context of a <InstallComponentProvider> component.]`
    )
  }
})

test('install component context should render with init state', () => {
  const { result } = renderHook(() => useInstallComponentContext(), {
    wrapper: ({ children }) => (
      <InstallComponentProvider initInstallComponent={mockInstallComponent}>
        {children}
      </InstallComponentProvider>
    ),
  })
  expect(result.current.installComponent).toHaveProperty(
    'id',
    mockInstallComponent.id
  )
  expect(result.current.installComponent).toHaveProperty(
    'component',
    mockInstallComponent.component
  )
  expect(result.current.isFetching).toBeFalsy()
})

test.skip(
  'installComponent context should refetch it state from api if provider has polling enabled',
  async () => {
    const { result } = renderHook(() => useInstallComponentContext(), {
      wrapper: ({ children }) => (
        <InstallComponentProvider
          initInstallComponent={mockInstallComponent}
          shouldPoll
        >
          {children}
        </InstallComponentProvider>
      ),
    })
    expect(result.current.installComponent).toHaveProperty(
      'id',
      mockInstallComponent.id
    )
    expect(result.current.installComponent).toHaveProperty(
      'component',
      mockInstallComponent.component
    )
    expect(result.current.isFetching).toBeFalsy()

    // starting installComponent refetch
    await waitFor(
      () => {
        return expect(result.current.isFetching).toBeTruthy()
      },
      {
        timeout: POLL_DURATION + 1000,
      }
    )

    // finish installComponent refetch
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
