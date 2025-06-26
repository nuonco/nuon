import '@test/mock-fetch-options'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import {
  getApps,
  getApp,
  getAppActionWorkflow,
  getAppActionWorkflows,
  getAppComponents,
  getAppConfigs,
  getAppInstalls,
  getAppLatestConfig,
  getAppLatestInputConfig,
  getAppLatestRunnerConfig,
  getAppLatestSandboxConfig,
} from './apps'

describe('getApps should handle response status codes from GET apps endpoint', () => {
  const orgId = 'test-id'
  test('200 status', async () => {
    const spec = await getApps({ orgId })
    spec.forEach((s) => {
      expect(s).toHaveProperty('id')
      expect(s).toHaveProperty('name')
      expect(s).toHaveProperty('cloud_platform')
    })
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getApps({ orgId }).catch((err) => expect(err).toMatchSnapshot())
  })
})

describe('getApp should handle response status codes from GET apps/:id endpoint', () => {
  const orgId = 'test-id'
  const appId = 'test-id'
  test('200 status', async () => {
    const spec = await getApp({ appId, orgId })
    expect(spec).toHaveProperty('id')
    expect(spec).toHaveProperty('name')
    expect(spec).toHaveProperty('cloud_platform')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getApp({ appId, orgId }).catch((err) => expect(err).toMatchSnapshot())
  })
})

describe('getAppComponents should handle response status codes from GET apps/:id/components endpoint', () => {
  const orgId = 'test-id'
  const appId = 'test-id'
  test('200 status', async () => {
    const spec = await getAppComponents({ appId, orgId })
    spec.forEach((s) => {
      expect(s).toHaveProperty('id')
      expect(s).toHaveProperty('name')
    })
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getAppComponents({ appId, orgId }).catch((err) =>
      expect(err).toMatchSnapshot()
    )
  })
})

describe('getAppConfigs should handle response status codes from GET apps/:id/configs endpoint', () => {
  const orgId = 'test-id'
  const appId = 'test-id'
  test('200 status', async () => {
    const spec = await getAppConfigs({ appId, orgId })
    spec.forEach((s) => {
      expect(s).toHaveProperty('id')
      expect(s).toHaveProperty('version')
    })
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getAppConfigs({ appId, orgId }).catch((err) =>
      expect(err).toMatchSnapshot()
    )
  })
})

describe('getAppInstalls should handle response status codes from GET apps/:id/installs endpoint', () => {
  const orgId = 'test-id'
  const appId = 'test-id'
  test('200 status', async () => {
    const spec = await getAppInstalls({ appId, orgId })
    spec.forEach((s) => {
      expect(s).toHaveProperty('id')
      expect(s).toHaveProperty('name')
    })
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getAppInstalls({ appId, orgId }).catch((err) =>
      expect(err).toMatchSnapshot()
    )
  })
})

describe('getAppLatestConfig should handle response status codes from GET apps/:id/latest-config endpoint', () => {
  const orgId = 'test-id'
  const appId = 'test-id'
  test('200 status', async () => {
    const spec = await getAppLatestConfig({ appId, orgId })
    expect(spec).toHaveProperty('id')
    expect(spec).toHaveProperty('version')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getAppLatestConfig({ appId, orgId }).catch((err) =>
      expect(err).toMatchSnapshot()
    )
  })
})

describe('getAppLatestInputConfig should handle response status codes from GET apps/:id/input-latest-config endpoint', () => {
  const orgId = 'test-id'
  const appId = 'test-id'
  test('200 status', async () => {
    const spec = await getAppLatestInputConfig({ appId, orgId })
    expect(spec).toHaveProperty('id')
    expect(spec).toHaveProperty('inputs')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getAppLatestInputConfig({ appId, orgId }).catch((err) =>
      expect(err).toMatchSnapshot()
    )
  })
})

describe('getAppLatestRunnerConfig should handle response status codes from GET apps/:id/runner-latest-config endpoint', () => {
  const orgId = 'test-id'
  const appId = 'test-id'
  test('200 status', async () => {
    const spec = await getAppLatestRunnerConfig({ appId, orgId })
    expect(spec).toHaveProperty('id')
    expect(spec).toHaveProperty('app_runner_type')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getAppLatestRunnerConfig({ appId, orgId }).catch((err) =>
      expect(err).toMatchSnapshot()
    )
  })
})

describe('getAppLatestSandboxConfig should handle response status codes from GET apps/:id/sandbox-latest-config endpoint', () => {
  const orgId = 'test-id'
  const appId = 'test-id'
  test('200 status', async () => {
    const spec = await getAppLatestSandboxConfig({ appId, orgId })
    expect(spec).toHaveProperty('id')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getAppLatestSandboxConfig({ appId, orgId }).catch((err) =>
      expect(err).toMatchSnapshot()
    )
  })
})

describe('getAppActionWorkflows should handle response status codes from GET apps/:id/action-workflows endpoint', () => {
  const orgId = 'test-id'
  const appId = 'test-id'
  test('200 status', async () => {
    const spec = await getAppActionWorkflows({ appId, orgId })
    spec.forEach((s) => {
      expect(s).toHaveProperty('id')
      expect(s).toHaveProperty('configs')
    })
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getAppActionWorkflows({ appId, orgId }).catch((err) =>
      expect(err).toMatchSnapshot()
    )
  })
})

describe('getAppActionWorkflow should handle response status codes from GET apps/:id/action-workflows endpoint', () => {
  const orgId = 'test-id'
  const actionWorkflowId = 'test-id'
  test('200 status', async () => {
    const spec = await getAppActionWorkflow({ actionWorkflowId, orgId })
    expect(spec).toHaveProperty('id')
    expect(spec).toHaveProperty('configs')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getAppActionWorkflow({ actionWorkflowId, orgId }).catch((err) =>
      expect(err).toMatchSnapshot()
    )
  })
})
