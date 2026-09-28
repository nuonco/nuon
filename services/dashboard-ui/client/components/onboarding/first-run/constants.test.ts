import { expect, test } from 'bun:test'
import { appFileStubs, CLOUD_SANDBOX, TEST_CLOUDS } from './constants'

const stubsFor = (cloud: (typeof TEST_CLOUDS)[number]) =>
  Object.fromEntries(
    appFileStubs({ appName: 'acme-api', cloud, repo: 'acme/acme-api' }).map((f) => [f.name, f.snippet])
  )

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
