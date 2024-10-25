import '@test/mock-fetch-options'
import { expect, test } from 'vitest'
import { getRunnerJob } from './get-runner-job'

const orgId = 'org-id'
const jobId = 'job-id'

test('getRunnerJob should return a runner job object', async () => {
  const spec = await getRunnerJob({
    orgId,
    jobId,
  })

  expect(spec).toHaveProperty('id')
  expect(spec).toHaveProperty('created_at')
  expect(spec).toHaveProperty('updated_at')
})

test('getRunnerJob should throw an error when it can not find a runner job', async () => {
  try {
    await getRunnerJob({
      orgId,
      jobId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Failed to fetch runner job]`)
  }
})
