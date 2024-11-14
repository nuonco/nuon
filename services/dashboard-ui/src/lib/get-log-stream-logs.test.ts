import '@test/mock-fetch-options'
import { expect, test } from 'vitest'
import { getLogStreamLogs } from './get-log-stream-logs'

const orgId = 'org-id'
const logStreamId = 'log-stream-id'

test('getLogStreamLogs should return an array of log stream logs object', async () => {
  const spec = await getLogStreamLogs({
    orgId,
    logStreamId,
  })

  expect(spec).toHaveLength(3)
  spec.map((s) => {
    expect(s).toHaveProperty('timestamp')
    expect(s).toHaveProperty('severity_number')
    expect(s).toHaveProperty('severity_text')
    expect(s).toHaveProperty('body')
  })
})

test('getLogStreamLogs should throw an error when it can not find any log stream logs', async () => {
  try {
    await getLogStreamLogs({
      orgId,
      logStreamId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Failed to fetch log stream logs]`)
  }
})
