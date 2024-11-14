import '@test/mock-fetch-options'
import { expect, test } from 'vitest'
import { getLogStream } from './get-log-stream'

const orgId = 'org-id'
const logStreamId = 'log-stream-id'

test('getLogStream should return a log stream object', async () => {
  const spec = await getLogStream({
    orgId,
    logStreamId,
  })

  expect(spec).toHaveProperty('id')
  expect(spec).toHaveProperty('created_at')
  expect(spec).toHaveProperty('updated_at')
})

test('getLogStream should throw an error when it can not find a log stream', async () => {
  try {
    await getLogStream({
      orgId,
      logStreamId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Failed to fetch log stream]`)
  }
})
