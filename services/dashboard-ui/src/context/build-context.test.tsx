import { afterAll, expect, test, vi } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { getGetBuild200Response } from '@test/mock-api-handlers'
import { useBuildContext, BuildProvider } from './build-context'

const mockBuild = getGetBuild200Response()
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

test('useBuildContext should throw error when used outside of BuildProvider', () => {
  try {
    renderHook(() => useBuildContext())
  } catch (error) {
    expect(error).toMatchInlineSnapshot(
      `[Error: useBuildContext() may only be used in the context of a <BuildProvider> component.]`
    )
  }
})

test('build context should render with init state', () => {
  const { result } = renderHook(() => useBuildContext(), {
    wrapper: ({ children }) => (
      <BuildProvider initBuild={mockBuild}>{children}</BuildProvider>
    ),
  })
  expect(result.current.build).toHaveProperty('id', mockBuild.id)
  expect(result.current.build).toHaveProperty(
    'component_id',
    mockBuild.component_id
  )
  expect(result.current.isFetching).toBeFalsy()
})

test(
  'build context should refetch it state from api if provider has polling enabled',
  async () => {
    const { result } = renderHook(() => useBuildContext(), {
      wrapper: ({ children }) => (
        <BuildProvider initBuild={{ ...mockBuild, org_id: 'test' }} shouldPoll>
          {children}
        </BuildProvider>
      ),
    })
    expect(result.current.build).toHaveProperty('id', mockBuild.id)
    expect(result.current.build).toHaveProperty(
      'component_id',
      mockBuild.component_id
    )
    expect(result.current.isFetching).toBeFalsy()

    // starting build refetch
    await waitFor(
      () => {
        return expect(result.current.isFetching).toBeTruthy()
      },
      {
        timeout: POLL_DURATION + 1000,
      }
    )

    // finish build refetch
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
