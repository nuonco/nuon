import { describe, expect, test } from 'bun:test'
import { buildCreateInstallBody } from './buildCreateInstallBody'
import type { InstallFormValues } from './schema'

const base: InstallFormValues = {
  name: '  my-install  ',
  region: 'us-west-2',
  aws_connection_id: '',
  aws_account_id: '',
  location: '',
  azure_subscription_id: '',
  gcp_project_id: '',
  autoApprove: true,
  vpcTemplateUrl: '',
  runnerTemplateUrl: '',
  labels: [],
  role: '',
  deployDependents: true,
  stackOnly: false,
  inputsOnly: false,
  inputs: { db_url: 'postgres://x', enable_tls: true, replicas: '3' },
}

describe('buildCreateInstallBody', () => {
  test('maps core fields with trimmed name and bool→string inputs', () => {
    const body = buildCreateInstallBody(base, 'aws')
    expect(body.name).toBe('my-install')
    expect(body.install_config?.approval_option).toBe('approve-all')
    expect(body.inputs).toEqual({
      db_url: 'postgres://x',
      enable_tls: 'true',
      replicas: '3',
    })
    expect(body.metadata?.managed_by).toBe('nuon/dashboard')
  })

  test('aws account carries region + optional connection/account', () => {
    const body = buildCreateInstallBody(
      { ...base, aws_connection_id: 'conn-1', aws_account_id: '123456789012' },
      'aws'
    )
    expect(body.aws_account).toEqual({
      iam_role_arn: '',
      region: 'us-west-2',
      connection_id: 'conn-1',
      account_id: '123456789012',
    })
  })

  test('prompt approval when autoApprove is false', () => {
    const body = buildCreateInstallBody({ ...base, autoApprove: false }, 'aws')
    expect(body.install_config?.approval_option).toBe('prompt')
  })

  test('includes advanced template overrides when set', () => {
    const body = buildCreateInstallBody(
      { ...base, vpcTemplateUrl: ' https://vpc ', runnerTemplateUrl: '' },
      'aws'
    )
    expect(body.install_config?.vpc_nested_template_url).toBe('https://vpc')
    expect(body.install_config?.runner_nested_template_url).toBeUndefined()
  })

  test('collects non-empty labels, dropping blank keys', () => {
    const body = buildCreateInstallBody(
      {
        ...base,
        labels: [
          { key: ' env ', value: ' prod ' },
          { key: '', value: 'ignored' },
        ],
      },
      'aws'
    )
    expect(body.labels).toEqual({ env: 'prod' })
  })

  test('azure and gcp account mapping', () => {
    const azure = buildCreateInstallBody(
      { ...base, region: '', location: 'eastus', azure_subscription_id: 'sub-1' },
      'azure'
    )
    expect(azure.azure_account?.location).toBe('eastus')
    expect(azure.azure_account?.subscription_id).toBe('sub-1')
    expect(azure.aws_account).toBeUndefined()

    const gcp = buildCreateInstallBody(
      { ...base, region: '', gcp_project_id: 'proj-1' },
      'gcp'
    )
    expect(gcp.gcp_account?.project_id).toBe('proj-1')
  })

  test('omits inputs/labels when empty', () => {
    const body = buildCreateInstallBody(
      { ...base, inputs: {}, labels: [] },
      'aws'
    )
    expect(body.inputs).toBeUndefined()
    expect(body.labels).toBeUndefined()
  })

  test('sends stack_only only when the stack-only scope is chosen', () => {
    expect(buildCreateInstallBody(base, 'aws').stack_only).toBeUndefined()
    expect(
      buildCreateInstallBody({ ...base, stackOnly: true }, 'aws').stack_only
    ).toBe(true)
  })
})
