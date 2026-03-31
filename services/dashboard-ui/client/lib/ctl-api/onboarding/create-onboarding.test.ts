import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { createOnboarding } from './create-onboarding'

describe('createOnboarding should handle response status codes from POST onboarding endpoint', () => {
  test('201 status', async () => {
    const result = await createOnboarding()
    expect(result).toHaveProperty('id')
    expect(result).toHaveProperty('status')
    expect(result).toHaveProperty('current_step')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(createOnboarding()).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
