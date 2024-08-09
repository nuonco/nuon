import '@test/mock-fetch-options'
import { describe, expect, test } from 'vitest'
import { getAppRunnerLatestConfig } from './get-app-runner-latest-config'

const appId = 'app-id'
const orgId = 'org-id'

describe('getAppRunnerLatestConfig', () => {
  test('should return the latest app runner config object', async () => {
    const spec = await getAppRunnerLatestConfig({
      appId,
      orgId,
    })

    expect(spec).toHaveProperty('id')
    expect(spec).toHaveProperty('created_at')
    expect(spec).toHaveProperty('updated_at')
    expect(spec).toHaveProperty('app_runner_type')
  })

  test('should throw an error when it can not find latest runner config for the app', async () => {
    try {
      await getAppRunnerLatestConfig({
        appId,
        orgId,
      })
    } catch (error) {
      expect(error).toMatchInlineSnapshot(
        `[Error: Failed to fetch latest app runner config]`
      )
    }
  })
})
