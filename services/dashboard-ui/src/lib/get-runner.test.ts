import '@test/mock-fetch-options'
import { expect, test } from 'vitest'
import { getRunner } from './get-runner'

const orgId = 'org-id'
const runnerId = 'job-id'

test('getRunner should return a runner object', async () => {
  const spec = await getRunner({
    orgId,
    runnerId,
  })

  expect(spec).toHaveProperty('id')
  expect(spec).toHaveProperty('created_at')
  expect(spec).toHaveProperty('updated_at')
})

test('getRunner should throw an error when it can not find a runner', async () => {
  try {
    await getRunner({
      orgId,
      runnerId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Failed to fetch runner]`)
  }
})
