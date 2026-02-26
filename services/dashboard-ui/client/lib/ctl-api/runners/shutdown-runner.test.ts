import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { shutdownRunner } from './shutdown-runner'

describe('shutdownRunner should handle response status codes from POST runners/:id/graceful-shutdown endpoint', () => {
  const orgId = 'test-org-id'
  const runnerId = 'test-runner-id'

  test('200 status with graceful shutdown (explicit)', async () => {
    const { data } = await shutdownRunner({
      force: false,
      orgId,
      runnerId,
    })
    expect(data).toBe(true)
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await shutdownRunner({
      orgId,
      runnerId,
    })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})

describe('shutdownRunner should handle response status codes from POST runners/:id/force-shutdown endpoint', () => {
  const orgId = 'test-org-id'
  const runnerId = 'test-runner-id'

  test('200 status with force shutdown', async () => {
    const { data } = await shutdownRunner({
      force: true,
      orgId,
      runnerId,
    })
    expect(data).toBe(true)
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await shutdownRunner({
      orgId,
      runnerId,
    })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
