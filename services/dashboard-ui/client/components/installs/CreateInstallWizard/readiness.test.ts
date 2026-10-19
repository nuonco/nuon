import { describe, expect, test } from 'bun:test'
import type { TAppInputConfig } from '@/types'
import type { InstallFormValues } from '@/components/installs/forms/InstallForm'
import {
  assignmentBlocker,
  enrollmentBlocker,
  installDetailsBlocker,
  installInputsBlocker,
  labelsMatchGroup,
  mergeCarriedLabels,
} from './readiness'
import { createInstallWizardPath, createWizardStatuses } from './wizard-path'

const values = (overrides: Partial<InstallFormValues> = {}): InstallFormValues => ({
  name: 'acme-west',
  region: 'us-west-2',
  cloud_connection_id: '',
  aws_account_id: '',
  location: '',
  azure_subscription_id: '',
  gcp_project_id: '',
  autoApprove: false,
  vpcTemplateUrl: '',
  runnerTemplateUrl: '',
  labels: [],
  role: '',
  deployDependents: true,
  stackOnly: false,
  inputsOnly: false,
  inputs: {},
  ...overrides,
})

describe('create install wizard', () => {
  test('builds the branch create path', () => {
    expect(createInstallWizardPath('org-1', 'app-1', 'br-1')).toBe(
      '/org-1/apps/app-1/branches/br-1/installs/new'
    )
  })

  test('marks steps up to the frontier as done', () => {
    expect(createWizardStatuses(0)).toEqual([
      'current',
      'upcoming',
      'upcoming',
      'upcoming',
      'upcoming',
    ])
    expect(createWizardStatuses(2)[0]).toBe('done')
    expect(createWizardStatuses(2)[2]).toBe('current')
    expect(createWizardStatuses(4).at(-1)).toBe('current')
  })

  test('requires a group when the branch has groups', () => {
    expect(enrollmentBlocker(0, { mode: 'none', group: null })).toBeUndefined()
    expect(enrollmentBlocker(2, { mode: 'none', group: null })).toBe(
      'Choose an install group'
    )
    expect(
      enrollmentBlocker(2, {
        mode: 'default',
        group: { id: 'g-1', name: 'Remaining installs', default: true },
      })
    ).toBeUndefined()
  })

  test('carries selected group labels without overwriting user labels', () => {
    expect(
      mergeCarriedLabels({
        labels: [
          { key: 'owner', value: 'payments' },
          { key: 'environment', value: 'staging' },
        ],
        previous: { environment: 'staging' },
        next: { environment: 'production', region: 'us-east-1' },
      })
    ).toEqual([
      { key: 'owner', value: 'payments' },
      { key: 'environment', value: 'production' },
      { key: 'region', value: 'us-east-1' },
    ])
  })

  test('prevents combining pinning with selector labels', () => {
    const group = {
      id: 'group-production',
      name: 'production',
      label_selector: { match_labels: { environment: 'production' } },
    }
    const config = { install_groups: [group] }
    const labels = [{ key: 'environment', value: 'production' }]

    expect(labelsMatchGroup(labels, group)).toBe(true)
    expect(
      assignmentBlocker({ mode: 'explicit', group }, labels, config)
    ).toBe('Remove the group selector labels or use label-based assignment')
    expect(
      assignmentBlocker({ mode: 'labels', group }, [], config)
    ).toBe('Restore the selected group labels or choose pin to group')
    expect(
      assignmentBlocker({ mode: 'labels', group }, labels, config)
    ).toBeUndefined()
  })

  test('gates install details by platform', () => {
    expect(
      installDetailsBlocker({
        values: values({ name: '  ' }),
        platform: 'aws',
        requireTargetAccount: false,
      })
    ).toBe('Enter the install name and destination')
    expect(
      installDetailsBlocker({
        values: values({ region: '' }),
        platform: 'aws',
        requireTargetAccount: false,
      })
    ).toBe('Enter the install name and region')
    expect(
      installDetailsBlocker({
        values: values({ aws_account_id: '123' }),
        platform: 'aws',
        requireTargetAccount: false,
      })
    ).toBe('Enter a 12-digit AWS account ID')
    expect(
      installDetailsBlocker({
        values: values(),
        platform: 'aws',
        requireTargetAccount: false,
      })
    ).toBeUndefined()
    expect(
      installDetailsBlocker({
        values: values({ location: '' }),
        platform: 'azure',
        requireTargetAccount: true,
      })
    ).toBe('Enter the install name and location')
  })

  test('gates required inputs', () => {
    const inputConfig = {
      input_groups: [
        {
          app_inputs: [
            { name: 'environment', required: true, source: 'vendor' },
            { name: 'replicas', required: false, source: 'vendor' },
          ],
        },
      ],
    } as TAppInputConfig

    expect(installInputsBlocker(values(), inputConfig)).toBe(
      'Fill in the required inputs'
    )
    expect(
      installInputsBlocker(
        values({ inputs: { environment: 'production' } }),
        inputConfig
      )
    ).toBeUndefined()
  })
})
