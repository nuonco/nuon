import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { completeInstallStep } from './complete-install-step'

describe('completeInstallStep should handle response status codes from POST onboarding/current/steps/install endpoint', () => {
  const orgId = 'test-org-id'
  const body = { name: 'test-install', install_id: 'test-install-id' }

  test('200 status', async () => {
    const result = await completeInstallStep({ body, orgId })
    expect(result).toHaveProperty('id')
    expect(result).toHaveProperty('status')
    expect(result).toHaveProperty('current_step')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(completeInstallStep({ body, orgId })).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
