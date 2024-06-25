import { afterAll, expect, test, vi } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { getGetOrg200Response } from '@test/mock-api-handlers'
import { useOrgContext, OrgProvider } from './org-context'

const mockOrg = getGetOrg200Response()
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

test('useOrgContext should throw error when used outside of OrgProvider', () => {
  try {
    renderHook(() => useOrgContext())
  } catch (error) {
    expect(error).toMatchInlineSnapshot(
      `[Error: useOrgContext() may only be used in the context of a <OrgProvider> component.]`
    )
  }
})

test('org context should render with init state', () => {
  const { result } = renderHook(() => useOrgContext(), {
    wrapper: ({ children }) => (
      <OrgProvider initOrg={mockOrg}>{children}</OrgProvider>
    ),
  })
  expect(result.current.org).toHaveProperty('id', mockOrg.id)
  expect(result.current.org).toHaveProperty('name', mockOrg.name)
  expect(result.current.isFetching).toBeFalsy()
})

test.skip(
  'org context should refetch it state from api if provider has polling enabled',
  async () => {
    const { result } = renderHook(() => useOrgContext(), {
      wrapper: ({ children }) => (
        <OrgProvider initOrg={mockOrg} shouldPoll>
          {children}
        </OrgProvider>
      ),
    })
    expect(result.current.org).toHaveProperty('id', mockOrg.id)
    expect(result.current.org).toHaveProperty('name', mockOrg.name)
    expect(result.current.isFetching).toBeFalsy()

    // starting org refetch
    await waitFor(
      () => {
        return expect(result.current.isFetching).toBeTruthy()
      },
      {
        timeout: POLL_DURATION + 1000,
      }
    )

    // finish org refetch
    await waitFor(
      () => {
        return expect(result.current.isFetching).toBeFalsy()
      },
      {
        timeout: POLL_DURATION + 1001,
      }
    )
  },
  POLL_DURATION * 1002
)
