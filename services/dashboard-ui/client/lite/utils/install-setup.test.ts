import { describe, expect, test } from 'bun:test'
import { isWizardComplete } from './wizard'
import {
  installSetupDescriptor,
  installSetupStateFromInstall,
} from './install-setup'

describe('install setup completeness', () => {
  test('is incomplete while provisioning', () => {
    expect(
      isWizardComplete(
        installSetupDescriptor,
        installSetupStateFromInstall({
          lifecycle_phase: { phase: 'provisioning' },
        })
      )
    ).toBe(false)
  })

  test('is complete once provisioned or after first provision', () => {
    expect(
      isWizardComplete(
        installSetupDescriptor,
        installSetupStateFromInstall({
          lifecycle_phase: { phase: 'provisioned' },
        })
      )
    ).toBe(true)
    expect(
      isWizardComplete(
        installSetupDescriptor,
        installSetupStateFromInstall({
          lifecycle_phase: { phase: 'deprovisioned' },
        })
      )
    ).toBe(true)
  })

  test('treats a missing phase as already through setup', () => {
    expect(
      isWizardComplete(installSetupDescriptor, installSetupStateFromInstall({}))
    ).toBe(true)
  })
})
