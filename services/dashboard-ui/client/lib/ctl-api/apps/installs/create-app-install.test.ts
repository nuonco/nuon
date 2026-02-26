import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { createAppInstall } from './create-app-install'

describe('createAppInstall should handle response status codes from POST apps/:appId/installs endpoint', () => {
  const appId = 'test-app-id'
  const orgId = 'test-org-id'

  test('201 status with AWS account', async () => {
    const { data: install } = await createAppInstall({
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
    expect(install).toHaveProperty('id')
    expect(install).toHaveProperty('name')
    expect(install).toHaveProperty('app_id')
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await createAppInstall({
      appId,
      orgId,
      body: { name: 'test-install' },
    })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
