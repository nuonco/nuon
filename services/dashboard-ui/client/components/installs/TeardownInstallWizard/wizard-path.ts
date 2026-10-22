export const TEARDOWN_WIZARD_STEPS = [
  {
    id: 'review',
    navLabel: 'Review',
    title: 'Review teardown',
    description: 'Confirm what this teardown will remove.',
  },
  {
    id: 'confirm',
    navLabel: 'Confirm',
    title: 'Confirm',
    description: 'Confirm teardown before it starts.',
  },
  {
    id: 'teardown',
    navLabel: 'Teardown',
    title: 'Teardown',
    description:
      'The teardown workflow removes the components, sandbox, and runner from the cloud account.',
  },
] as const

export const teardownWizardStatuses = (farthest: number) =>
  TEARDOWN_WIZARD_STEPS.map((_, index) => {
    if (index < farthest) return 'done' as const
    if (index === farthest) return 'current' as const
    return 'upcoming' as const
  })
