import '@test/mock-fetch-options'
import { describe, expect, test } from 'vitest'
import { getAppConfigs } from './get-app-configs'

const appId = 'app-id'
const orgId = 'org-id'

describe('getAppConfigs', () => {
  test('should return an array of app configs object', async () => {
    const spec = await getAppConfigs({
      appId,
      orgId,
    })

    expect(spec).toHaveLength(3)
    spec.map((s) => {
      expect(s).toHaveProperty('id')
      expect(s).toHaveProperty('created_at')
      expect(s).toHaveProperty('updated_at')
      expect(s).toHaveProperty('state')
    })
  })

  test('should throw an error when it can not find any configs for the app', async () => {
    try {
      await getAppConfigs({
        appId,
        orgId,
      })
    } catch (error) {
      expect(error).toMatchInlineSnapshot(
        `[Error: Failed to fetch app configs]`
      )
    }
  })
})
