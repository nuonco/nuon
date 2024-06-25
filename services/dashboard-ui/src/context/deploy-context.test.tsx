import { afterAll, expect, test, vi } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { getGetInstallDeploy200Response } from '@test/mock-api-handlers'
import {
  useInstallDeployContext,
  InstallDeployProvider,
} from './deploy-context'

const mockInstallDeploy = getGetInstallDeploy200Response()
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

test('useInstallDeployContext should throw error when used outside of InstallDeployProvider', () => {
  try {
    renderHook(() => useInstallDeployContext())
  } catch (error) {
    expect(error).toMatchInlineSnapshot(
      `[Error: useInstallDeployContext() may only be used in the context of a <InstallDeployProvider> component.]`
    )
  }
})

test('deploy context should render with init state', () => {
  const { result } = renderHook(() => useInstallDeployContext(), {
    wrapper: ({ children }) => (
      <InstallDeployProvider initDeploy={mockInstallDeploy}>
        {children}
      </InstallDeployProvider>
    ),
  })
  expect(result.current.deploy).toHaveProperty('id', mockInstallDeploy.id)
  expect(result.current.deploy).toHaveProperty(
    'component_id',
    mockInstallDeploy.component_id
  )
  expect(result.current.isFetching).toBeFalsy()
})

test.skip(
  'deploy context should refetch it state from api if provider has polling enabled',
  async () => {
    const { result } = renderHook(() => useInstallDeployContext(), {
      wrapper: ({ children }) => (
        <InstallDeployProvider
          initDeploy={{ ...mockInstallDeploy, org_id: 'test' }}
          shouldPoll
        >
          {children}
        </InstallDeployProvider>
      ),
    })
    expect(result.current.deploy).toHaveProperty('id', mockInstallDeploy.id)
    expect(result.current.deploy).toHaveProperty(
      'component_id',
      mockInstallDeploy.component_id
    )
    expect(result.current.isFetching).toBeFalsy()

    // starting deploy refetch
    await waitFor(
      () => {
        return expect(result.current.isFetching).toBeTruthy()
      },
      {
        timeout: POLL_DURATION + 1000,
      }
    )

    // finish deploy refetch
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
