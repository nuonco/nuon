import '@test/mock-fetch-options'
import { describe, expect, test } from 'vitest'
import { getAppSandboxLatestConfig } from './get-app-sandbox-latest-config'

const appId = 'app-id'
const orgId = 'org-id'

describe('getAppSandboxLatestConfig', () => {
  test('should return the latest app inputs config object', async () => {
    const spec = await getAppSandboxLatestConfig({
      appId,
      orgId,
    })

    expect(spec).toHaveProperty('id')
    expect(spec).toHaveProperty('created_at')
    expect(spec).toHaveProperty('updated_at')
    expect(spec).toHaveProperty('cloud_platform')
  })

  test('should throw an error when it can not find latest sandbox config for the app', async () => {
    try {
      await getAppSandboxLatestConfig({
        appId,
        orgId,
      })
    } catch (error) {
      expect(error).toMatchInlineSnapshot(
        `[Error: Failed to fetch latest app sandbox config]`
      )
    }
  })
})
