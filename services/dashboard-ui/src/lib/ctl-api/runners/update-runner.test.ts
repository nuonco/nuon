import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { updateRunner, type IUpdateRunnerBody } from './update-runner'

describe('updateRunner should handle response status codes from PATCH runners/:id/settings endpoint', () => {
  const orgId = 'test-org-id'
  const runnerId = 'test-runner-id'

  test('200 status', async () => {
    const body: IUpdateRunnerBody = {
      container_image_tag: 'v1.0.0',
      container_image_url: 'registry.example.com/runner:v1.0.0',
      org_awsiam_role_arn: 'arn:aws:iam::123456789012:role/test-role',
      org_k8s_service_account_name: 'test-service-account',
      runner_api_url: 'https://api.example.com/runner',
    }

    const { data } = await updateRunner({
      body,
      orgId,
      runnerId,
    })
    expect(data).toHaveProperty('id')
    expect(data).toHaveProperty('status')
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const body: IUpdateRunnerBody = {
      container_image_tag: 'v1.0.0',
      container_image_url: 'registry.example.com/runner:v1.0.0',
      org_awsiam_role_arn: 'arn:aws:iam::123456789012:role/test-role',
      org_k8s_service_account_name: 'test-service-account',
      runner_api_url: 'https://api.example.com/runner',
    }

    const { error, status } = await updateRunner({
      body,
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