export const INSTALL_SHARE_SCOPES = [
  'approvals',
  'readme',
  'operations',
  'resources',
] as const

export type TInstallShareScope = (typeof INSTALL_SHARE_SCOPES)[number]

export type TInstallShareStatus = 'pending' | 'accepted' | 'revoked'

export type TInstallShare = {
  id: string
  email: string
  scopes: TInstallShareScope[]
  status: TInstallShareStatus
  shared_at: string
  install_id?: string
  install_name?: string
  app_name?: string
}

export type TShareInstallValues = {
  email: string
  approvals: boolean
  readme: boolean
  operations: boolean
  resources: boolean
}

export const SHARE_SCOPE_OPTIONS: {
  key: TInstallShareScope
  label: string
  description: string
}[] = [
  {
    key: 'approvals',
    label: 'Approvals',
    description: 'Review and respond to pending deployment plans.',
  },
  {
    key: 'readme',
    label: 'README',
    description: 'View the install README.',
  },
  {
    key: 'operations',
    label: 'Operations',
    description: 'View operations activity for this install.',
  },
  {
    key: 'resources',
    label: 'Resources',
    description: 'View install resources and alignment.',
  },
]

export const scopeLabels = (scopes: TInstallShareScope[] = []) =>
  scopes.map(
    (scope) =>
      SHARE_SCOPE_OPTIONS.find((option) => option.key === scope)?.label ?? scope
  )

export const scopesFromValues = (
  values: Pick<
    TShareInstallValues,
    'approvals' | 'readme' | 'operations' | 'resources'
  >
): TInstallShareScope[] =>
  INSTALL_SHARE_SCOPES.filter((scope) => values[scope])
