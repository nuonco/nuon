import { describe, expect, test } from 'bun:test'
import type { TAppBranch, TAppBranchInstallGroup } from '@/types'
import { isWizardComplete } from './wizard'
import { appSetupDescriptor, appSetupStateFromBranches } from './app-setup'

const branch = (groups: TAppBranchInstallGroup[]): TAppBranch => ({
  id: 'br_main',
  configs: [{ config_number: 1, install_groups: groups }],
})

describe('app setup completeness', () => {
  test('is incomplete without a deployment plan', () => {
    expect(
      isWizardComplete(appSetupDescriptor, appSetupStateFromBranches([]))
    ).toBe(false)
    expect(
      isWizardComplete(
        appSetupDescriptor,
        appSetupStateFromBranches([branch([])])
      )
    ).toBe(false)
    expect(
      isWizardComplete(
        appSetupDescriptor,
        appSetupStateFromBranches([branch([{ name: 'manual' }])])
      )
    ).toBe(false)
  })

  test('is complete with a default or label-scoped group', () => {
    expect(
      isWizardComplete(
        appSetupDescriptor,
        appSetupStateFromBranches([
          branch([{ name: 'default', default: true }]),
        ])
      )
    ).toBe(true)
    expect(
      isWizardComplete(
        appSetupDescriptor,
        appSetupStateFromBranches([
          branch([
            {
              name: 'env',
              label_selector: { match_labels: { env: 'production' } },
            },
          ]),
        ])
      )
    ).toBe(true)
  })
})
