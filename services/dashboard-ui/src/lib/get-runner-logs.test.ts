import '@test/mock-fetch-options'
import { expect, test } from 'vitest'
import { getRunnerLogs } from './get-runner-logs'

const orgId = 'org-id'
const runnerId = 'runner-id'
const jobId = 'job-id'

test.skip('getRunnerLogs should return an array of runner log object', async () => {
  const spec = await getRunnerLogs({
    orgId,
    runnerId,
    jobId,
  })

  expect(spec).toHaveLength(3)
  spec.map((s) => {
    expect(s).toHaveProperty('timestamp')
    expect(s).toHaveProperty('severity_number')
    expect(s).toHaveProperty('severity_text')
    expect(s).toHaveProperty('body')
  })
})

test.skip('getRunnerLogs should throw an error when it can not find any runner logs', async () => {
  try {
    await getRunnerLogs({
      orgId,
      runnerId,
      jobId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Failed to fetch runner logs]`)
  }
})
