import '@test/mock-fetch-options'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import {
  cancelRunnerJob,
  getLogStream,
  getRunner,
  getRunnerJob,
  getRunnerJobs,
  getRunnerHealthChecks,
  getRunnerLatestHeartbeat,
} from './runners'

describe('getLogStream should handle response status codes from GET log-streams/:id endpoint', () => {
  const orgId = 'test-id'
  const logStreamId = 'test-id'
  test('200 status', async () => {
    const spec = await getLogStream({ logStreamId, orgId })
    expect(spec).toHaveProperty('id')
    expect(spec).toHaveProperty('open')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getLogStream({ logStreamId, orgId })
      .then((err) => expect(err).toMatchSnapshot())
      .catch((err) => expect(err).toMatchSnapshot())
  })
})

describe('getRunner should handle response status codes from GET runners/:id endpoint', () => {
  const orgId = 'test-id'
  const runnerId = 'test-id'
  test('200 status', async () => {
    const spec = await getRunner({ runnerId, orgId })
    expect(spec).toHaveProperty('id')
    expect(spec).toHaveProperty('display_name')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getRunner({ runnerId, orgId })
      .then((err) => expect(err).toMatchSnapshot())
      .catch((err) => expect(err).toMatchSnapshot())
  })
})

describe('getRunnerJob should handle response status codes from GET runner-jobs/:id endpoint', () => {
  const orgId = 'test-id'
  const runnerJobId = 'test-id'
  test('200 status', async () => {
    const spec = await getRunnerJob({ runnerJobId, orgId })
    expect(spec).toHaveProperty('id')
    expect(spec).toHaveProperty('executions')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getRunnerJob({ runnerJobId, orgId })
      .then((err) => expect(err).toMatchSnapshot())
      .catch((err) => expect(err).toMatchSnapshot())
  })
})

describe.skip('getRunnerJobs should handle response status codes from GET runners/:id/jobs endpoint', () => {
  const orgId = 'test-id'
  const runnerId = 'test-id'
  test('200 status', async () => {
    const { runnerJobs: spec } = await getRunnerJobs({ runnerId, orgId })
    spec.forEach((s) => {
      expect(s).toHaveProperty('id')
      expect(s).toHaveProperty('executions')
    })
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getRunnerJobs({ runnerId, orgId })
      .then((err) => expect(err).toMatchSnapshot())
      .catch((err) => expect(err).toMatchSnapshot())
  })
})

describe('cancelRunnerJob should handle response status codes from POST runner-jobs/:id/cancel endpoint', () => {
  const orgId = 'test-id'
  const runnerJobId = 'test-id'
  test('200 status', async () => {
    const spec = await cancelRunnerJob({ runnerJobId, orgId })
    expect(spec).toHaveProperty('id')
    expect(spec).toHaveProperty('executions')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await cancelRunnerJob({ runnerJobId, orgId }).catch((err) =>
      expect(err).toMatchSnapshot()
    )
  })
})

describe('getRunnerHealthChecks should handle response status codes from GET runners/:id/recent-health-checks endpoint', () => {
  const orgId = 'test-id'
  const runnerId = 'test-id'
  test('200 status', async () => {
    const spec = await getRunnerHealthChecks({ runnerId, orgId })
    spec.map((s) => {
      expect(s).toHaveProperty('minute_bucket')
      expect(s).toHaveProperty('status_code')
      expect(s).toHaveProperty('runner_id')
    })
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getRunnerHealthChecks({ runnerId, orgId })
      .then((err) => expect(err).toMatchSnapshot())
      .catch((err) => expect(err).toMatchSnapshot())
  })
})

describe('getRunnerLatestHeartbeat should handle response status codes from GET runners/:id/latest-heart-beat endpoint', () => {
  const orgId = 'test-id'
  const runnerId = 'test-id'
  test('200 status', async () => {
    const spec = await getRunnerLatestHeartbeat({ runnerId, orgId })
    expect(spec).toHaveProperty('id')
    expect(spec).toHaveProperty('alive_time')
    expect(spec).toHaveProperty('version')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getRunnerLatestHeartbeat({ runnerId, orgId })
      .then((err) => expect(err).toMatchSnapshot())
      .catch((err) => expect(err).toMatchSnapshot())
  })
})
