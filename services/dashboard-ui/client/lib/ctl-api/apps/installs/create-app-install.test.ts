import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { createAppInstall } from './create-app-install'

describe('createAppInstall should handle response status codes from POST apps/:appId/installs endpoint', () => {
  const appId = 'test-app-id'
  const orgId = 'test-org-id'

  test('201 status with AWS account', async () => {
    const result = await createAppInstall({
      appId,
      orgId,
      body: {
        name: 'test-aws-install',
        aws_account: {
          iam_role_arn: '',
          region: 'us-east-1',
        },
      },
    })
    expect(result).toHaveProperty('id')
    expect(result).toHaveProperty('name')
    expect(result).toHaveProperty('app_id')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(
      createAppInstall({ appId, orgId, body: { name: 'test-install' } })
    ).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
