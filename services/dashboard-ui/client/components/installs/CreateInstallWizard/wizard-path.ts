export const createInstallWizardPath = (
  orgId: string,
  appId: string,
  branchId: string
) => `/${orgId}/apps/${appId}/branches/${branchId}/installs/new`

export const CREATE_WIZARD_STEPS = [
  {
    id: 'app',
    navLabel: 'App',
    title: 'Select app',
    description:
      'This install is created from the current app. Choose the group it joins.',
  },
  {
    id: 'details',
    navLabel: 'Install details',
    title: 'Configure install',
    description:
      'Set the install identity, destination, approval behavior, and initial provisioning scope.',
  },
  {
    id: 'inputs',
    navLabel: 'Inputs',
    title: 'Configure inputs',
    description: 'Set the app inputs for this install.',
  },
  {
    id: 'review',
    navLabel: 'Review',
    title: 'Review',
    description: 'Confirm the install before creating it.',
  },
  {
    id: 'provision',
    navLabel: 'Provision',
    title: 'Provision',
    description:
      'The provision workflow builds the runner, sandbox, and components in the cloud account.',
  },
] as const

export type TCreateWizardStepId = (typeof CREATE_WIZARD_STEPS)[number]['id']

export type TCreateWizardStepStatus = 'done' | 'current' | 'upcoming'

export const createWizardStatuses = (
  farthest: number
): TCreateWizardStepStatus[] =>
  CREATE_WIZARD_STEPS.map((_, index) => {
    if (index < farthest) return 'done'
    if (index === farthest) return 'current'
    return 'upcoming'
  })
