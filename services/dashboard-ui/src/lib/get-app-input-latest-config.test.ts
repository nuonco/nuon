import '@test/mock-fetch-options'
import { describe, expect, test } from 'vitest'
import { getAppInputLatestConfig } from './get-app-input-latest-config'

const appId = 'app-id'
const orgId = 'org-id'

describe('getAppInputLatestConfig', () => {
  test('should return the latest app inputs config object', async () => {
    const spec = await getAppInputLatestConfig({
      appId,
      orgId,
    })

    expect(spec).toHaveProperty('id')
    expect(spec).toHaveProperty('created_at')
    expect(spec).toHaveProperty('updated_at')
    expect(spec).toHaveProperty('inputs')
  })

  test('should throw an error when it can not find latest input config for the app', async () => {
    try {
      await getAppInputLatestConfig({
        appId,
        orgId,
      })
    } catch (error) {
      expect(error).toMatchInlineSnapshot(
        `[Error: Failed to fetch latest app input config]`
      )
    }
  })
})
