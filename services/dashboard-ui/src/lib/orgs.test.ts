import '@test/mock-fetch-options'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getOrg, getOrgs, getVCSConnections, postOrg, nueGetOrg } from './orgs'

describe('getOrgs should handle response status codes from GET orgs endpoint', () => {
  test('200 status', async () => {
    const spec = await getOrgs()
    spec.forEach((s) => {
      expect(s).toHaveProperty('id')
      expect(s).toHaveProperty('name')
    })
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getOrgs().catch((err) => expect(err).toMatchSnapshot())
  })
})

describe('getOrg should handle response status codes from GET orgs/current endpoint', () => {
  const orgId = 'test-id'
  test('200 status', async () => {
    const spec = await getOrg({ orgId })
    expect(spec).toHaveProperty('id')
    expect(spec).toHaveProperty('name')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getOrg({ orgId }).catch((err) => expect(err).toMatchSnapshot())
  })
})

describe('getVCSConnections should handle response status codes from GET vcs/connections endpoint', () => {
  const orgId = 'test-id'
  test('200 status', async () => {
    const spec = await getVCSConnections({ orgId })
    spec.forEach((s) => {
      expect(s).toHaveProperty('id')
      expect(s).toHaveProperty('github_install_id')
    })
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getVCSConnections({ orgId }).catch((err) =>
      expect(err).toMatchSnapshot()
    )
  })
})

describe('postOrg should handle response status codes from POST orgs endpoint', () => {
  test('200 status', async () => {
    const spec = await postOrg({ name: 'test-name' })
    expect(spec).toHaveProperty('id')
    expect(spec).toHaveProperty('name')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await postOrg({ name: 'test-name' }).catch((err) =>
      expect(err).toMatchSnapshot()
    )
  })
})


describe('nueGetOrg should handle response status codes from GET orgs/current endpoint', () => {
  const orgId = 'test-id'
  test('200 status', async () => {
    const { data, error, status } = await nueGetOrg({ orgId })
    expect(data).toHaveProperty("id")
    expect(data).toHaveProperty("name")
    expect(error).toBeNull()
    expect(status).toBe(200)
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { data, error, status } = await nueGetOrg({ orgId })
    expect(data).toBeNull()
    expect(error).toHaveProperty("description")
    expect(error).toHaveProperty("error")
    expect(error).toHaveProperty("user_error")
    expect(status).toBe(code)
  })
})
