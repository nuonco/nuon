import { describe, expect, test } from 'vitest'
import { shutdownRunner } from './shutdown-runner'

describe('shutdownRunner should handle response status codes from POST runners/:id/graceful-shutdown endpoint', () => {
  const orgId = 'test-org-id'
  const runnerId = 'test-runner-id'

  test('200 status with graceful shutdown (explicit)', async () => {
    const result = await shutdownRunner({ force: false, orgId, runnerId })
    expect(result).toBe(true)
  })
})

describe('shutdownRunner should handle response status codes from POST runners/:id/force-shutdown endpoint', () => {
  const orgId = 'test-org-id'
  const runnerId = 'test-runner-id'

  test('200 status with force shutdown', async () => {
    const result = await shutdownRunner({ force: true, orgId, runnerId })
    expect(result).toBe(true)
  })
})
