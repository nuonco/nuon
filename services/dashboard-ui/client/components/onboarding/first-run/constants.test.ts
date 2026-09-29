import { expect, test } from 'bun:test'
import { appFileStubs, CLOUD_SANDBOX, TEST_CLOUDS, defaultRegion, regionOptions } from './constants'

const stubsFor = (cloud: (typeof TEST_CLOUDS)[number]) =>
  Object.fromEntries(
    appFileStubs({ appName: 'acme-api', cloud, repo: 'acme/acme-api' }).map((f) => [f.name, f.snippet])
  )

test('region options match the create-install catalogs', () => {
  expect(defaultRegion('aws')).toBe('us-east-1')
  expect(defaultRegion('gcp')).toBe('us-central1')
  expect(defaultRegion('azure')).toBe('eastus')

  const aws = regionOptions('aws')
  const virginia = aws.find((option) => option.value === 'us-east-1')
  expect(aws.length).toBeGreaterThan(4)
  expect(virginia?.label).toContain('US East (N. Virginia)')
  expect(virginia?.label).toContain('[us-east-1]')

  const azure = regionOptions('azure')
  const eastUs = azure.find((option) => option.value === 'eastus')
  expect(azure.length).toBeGreaterThan(4)
  expect(eastUs?.label).toContain('East US')
  expect(eastUs?.label).not.toContain('[eastus]')

  const gcp = regionOptions('gcp')
  expect(gcp.find((option) => option.value === 'us-central1')?.label).toContain('[us-central1]')
})

test('sandbox repos per cloud', () => {
  expect(CLOUD_SANDBOX).toEqual({
    aws: 'nuonco/aws-eks-auto-sandbox',
    gcp: 'nuonco/gcp-gke-sandbox',
    azure: 'nuonco/azure-aks-sandbox',
  })
})

for (const cloud of TEST_CLOUDS) {
  test(`${cloud} stubs carry its runner type, sandbox repo, and permission key`, () => {
    const stubs = stubsFor(cloud)
    expect(stubs['runner.toml']).toContain(`runner_type = "${cloud}"`)
    expect(stubs['sandbox.toml']).toContain(`repo      = "${CLOUD_SANDBOX[cloud]}"`)
    const key = { aws: 'managed_policy_name', gcp: 'gcp_predefined_role', azure: 'azure_built_in_roles' }[cloud]
    expect(stubs['permissions.toml']).toContain(key)
    expect(stubs).toMatchSnapshot()
  })
}
