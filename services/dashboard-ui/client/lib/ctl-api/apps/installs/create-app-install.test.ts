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
    expect(result).toHaveProperty('data')
    expect(result.data).toHaveProperty('id')
    expect(result.data).toHaveProperty('name')
    expect(result.data).toHaveProperty('app_id')
  })
})
