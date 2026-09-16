import { describe, expect, test } from 'bun:test'
import {
  buildCreateInstallBody,
  isDuplicateInstallNameError,
  normalizeInstallPlatform,
  pickInstallConfig,
} from './create-install'

const values = {
  name: '  production  ',
  location: 'us-west-2',
  accountId: '123456789012',
  autoApprove: true,
  stackOnly: true,
  labels: [
    { key: 'team', value: ' payments ' },
    { key: '', value: '' },
  ],
  inputs: [
    { name: 'hostname', value: 'payments.example.com' },
    { name: 'metrics', value: true },
  ],
  branchId: 'branch_main',
  installGroupId: 'group_production',
}

describe('create install helpers', () => {
  test('normalizes runner types to cloud platforms', () => {
    expect(normalizeInstallPlatform('aws-eks')).toBe('aws')
    expect(normalizeInstallPlatform('azure-aks')).toBe('azure')
    expect(normalizeInstallPlatform('gcp-gke')).toBe('gcp')
    expect(normalizeInstallPlatform(undefined)).toBe('unknown')
  })

  test('selects an active non-preview config for the requested scope', () => {
    const configs = [
      { id: 'preview', labels: { source: 'git-preview-run' } },
      { id: 'old', status: 'outdated' as const },
      { id: 'unbranched', status: 'active' as const },
      {
        id: 'branched',
        status: 'active' as const,
        app_branch_id: 'branch_main',
      },
    ]

    expect(pickInstallConfig(configs)?.id).toBe('unbranched')
    expect(pickInstallConfig(configs, true)?.id).toBe('unbranched')
    expect(pickInstallConfig([configs[3]])?.id).toBe('branched')
  })

  test('builds the AWS create body and lets group labels win', () => {
    expect(
      buildCreateInstallBody({
        values,
        platform: 'aws',
        groupLabels: { team: 'platform', env: 'production' },
      })
    ).toEqual({
      name: 'production',
      inputs: {
        hostname: 'payments.example.com',
        metrics: 'true',
      },
      install_config: { approval_option: 'approve-all' },
      labels: { team: 'platform', env: 'production' },
      metadata: { managed_by: 'nuon/dashboard' },
      stack_only: true,
      app_branch_id: 'branch_main',
      aws_account: {
        account_id: '123456789012',
        iam_role_arn: '',
        region: 'us-west-2',
      },
    })
  })

  test('recognizes duplicate install name errors', () => {
    expect(
      isDuplicateInstallNameError({
        error: 'unable to create install: duplicated key not allowed',
        description: 'duplicate key',
        user_error: true,
        status: 409,
      })
    ).toBe(true)
    expect(isDuplicateInstallNameError(new Error('request failed'))).toBe(false)
  })
})
