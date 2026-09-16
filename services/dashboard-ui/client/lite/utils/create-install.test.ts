import { describe, expect, test } from 'bun:test'
import type { TAppConfig } from '@/types'
import {
  buildCreateInstallBody,
  installSetupInputs,
  isDuplicateInstallNameError,
  resolveInstallConfig,
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

  test('falls back to the newest branch config when nothing is unbranched', () => {
    const branched = [
      { id: 'cfg_new', status: 'active', app_branch_id: 'branch_main' },
      { id: 'cfg_old', status: 'active', app_branch_id: 'branch_stable' },
    ] as TAppConfig[]

    expect(resolveInstallConfig(branched)).toEqual({
      config: branched[0],
      branchId: 'branch_main',
    })

    const withUnbranched = [
      ...branched,
      { id: 'cfg_sync', status: 'active' },
    ] as TAppConfig[]

    expect(resolveInstallConfig(withUnbranched)).toEqual({
      config: withUnbranched[2],
      branchId: '',
    })

    expect(
      resolveInstallConfig(
        [
          { id: 'cfg_branch', status: 'active', app_branch_id: 'branch_main' },
        ] as TAppConfig[],
        'branch_main'
      )
    ).toEqual({
      config: {
        id: 'cfg_branch',
        status: 'active',
        app_branch_id: 'branch_main',
      },
      branchId: 'branch_main',
    })

    expect(
      resolveInstallConfig([{ id: 'cfg_bad', status: 'error' }] as TAppConfig[])
    ).toEqual({ config: undefined, branchId: '' })
  })

  test('reads inputs from the flat array the API populates', () => {
    expect(
      installSetupInputs({
        input_groups: [
          { id: 'group_advanced', index: 1, app_inputs: [] },
          { id: 'group_main', index: 0, app_inputs: [] },
          {
            id: 'group_overrides',
            name: 'nuon_component_overrides',
            index: 1_000_000,
            app_inputs: [],
          },
        ],
        inputs: [
          {
            name: 'replicas',
            display_name: 'Replica count',
            group_id: 'group_main',
            index: 1,
            type: 'number',
            default: '2',
          },
          {
            name: 'hostname',
            display_name: 'Public hostname',
            group_id: 'group_main',
            index: 0,
            required: true,
          },
          {
            name: 'metrics',
            group_id: 'group_advanced',
            type: 'bool',
            default: 'true',
          },
          {
            name: 'customer_only',
            group_id: 'group_main',
            source: 'customer',
          },
          {
            name: 'nuon_component_override_v1_helm_values_6162',
            group_id: 'group_overrides',
            type: 'yaml',
          },
        ],
      })
    ).toEqual([
      {
        name: 'hostname',
        label: 'Public hostname',
        description: undefined,
        required: true,
        type: 'text',
        defaultValue: '',
      },
      {
        name: 'replicas',
        label: 'Replica count',
        description: undefined,
        required: undefined,
        type: 'number',
        defaultValue: '2',
      },
      {
        name: 'metrics',
        label: 'metrics',
        description: undefined,
        required: undefined,
        type: 'boolean',
        defaultValue: true,
      },
    ])
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
