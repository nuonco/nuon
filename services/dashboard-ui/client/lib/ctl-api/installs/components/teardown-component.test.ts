import { describe, expect, test } from 'vitest'
import { teardownComponent } from './teardown-component'

describe('teardownComponent should handle response status codes from POST installs/:installId/components/:componentId/teardown endpoint', () => {
  const componentId = 'test-component-id'
  const installId = 'test-install-id'
  const orgId = 'test-org-id'

  test('200 status with default body', async () => {
    const result = await teardownComponent({
      body: { error_behavior: 'continue', plan_only: true },
      componentId,
      installId,
      orgId,
    })
    expect(result).toBeDefined()
  })
})
