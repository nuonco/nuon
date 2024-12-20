import '@test/mock-fetch-options'
import { expect, test } from 'vitest'
import { getRunnerJobs } from './get-runner-jobs'

const orgId = 'org-id'
const runnerId = 'runner-id'

test('getRunnerJobs should return a list of runner job objects', async () => {
  const spec = await getRunnerJobs({
    orgId,
    runnerId,
  })

  spec.forEach((s) => {
    expect(s).toHaveProperty('id')
    expect(s).toHaveProperty('created_at')
    expect(s).toHaveProperty('updated_at')
  })
})

test('getRunnerJobs should throw an error when it can not find any runner jobs', async () => {
  try {
    await getRunnerJobs({
      orgId,
      runnerId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(`[Error: Failed to fetch runner jobs]`)
  }
})
